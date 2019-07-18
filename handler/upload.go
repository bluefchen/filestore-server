package handler

import (
	"bytes"
	"encoding/json"
	cmn "filestore-server/common"
	cfg "filestore-server/config"
	dblayer "filestore-server/db"
	"filestore-server/meta"
	"filestore-server/mq"
	"filestore-server/store/ceph"
	"filestore-server/store/oss"
	"filestore-server/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func init() {
	// 目录已存在
	if _, err := os.Stat(cfg.TempLocalRootDir); err == nil {
		return
	}

	// 尝试创建目录
	err := os.MkdirAll(cfg.TempLocalRootDir, 0744)
	if err != nil {
		log.Println("无法创建临时存储目录，程序将退出")
		os.Exit(1)
	}
}

// UploadHandler ： 跳转文件上传页面
func UploadHandler(c *gin.Context) {
	data, err := ioutil.ReadFile("../../static/view/upload.html")
	if err != nil {
		c.String(404, `网页不存在`)
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, string(data))
	//c.Redirect(http.StatusFound, "/static/view/upload.html")
}

//DoUploadHandler：处理上传文件
func DoUploadHandler(c *gin.Context) {
	errCode := 0
	//返回前调用
	defer func() {
		if errCode < 0 {
			c.JSON(http.StatusOK, gin.H{
				"code": errCode,
				"msg":  "Upload Failed",
			})
		}
	}()

	// 1. 从form表单中获得文件内容句柄
	file, head, err := c.Request.FormFile("file")
	if err != nil {
		errCode = -1
		log.Println("get form data file err:", err.Error())
		return
	}
	defer file.Close()

	// 2. 把文件内容转为[]byte
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		errCode = -2
		log.Println("Failed to get file data, err:", err.Error())
		return
	}

	// 3. 构建文件元信息
	fileMeta := meta.FileMeta{
		FileName: head.Filename,
		FileSha1: util.Sha1(buf.Bytes()), //　计算文件sha1
		FileSize: int64(len(buf.Bytes())),
		UploadAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	// 4. 将文件写入临时存储位置
	fileMeta.Location = cfg.TempLocalRootDir + fileMeta.FileName // 临时存储地址
	newFile, err := os.Create(fileMeta.Location)
	if err != nil {
		log.Println("Failed to create file, err:", err.Error())
		errCode = -3
		return
	}
	defer newFile.Close()

	nByte, err := newFile.Write(buf.Bytes())
	// 丢失数据
	if int64(nByte) != fileMeta.FileSize || err != nil {
		errCode = -4
		log.Println("Failed to save data into file, err:", err.Error())
		return
	}

	// 5. 同步或异步将文件转移到Ceph/OSS
	newFile.Seek(0, 0) // 游标重新回到文件头部
	if cfg.CurrentStoreType == cmn.StoreCeph {
		// 文件写入Ceph存储
		data, _ := ioutil.ReadAll(newFile)
		cephPath := "/ceph/" + fileMeta.FileSha1
		_ = ceph.PutObject("userfile", cephPath, data)
		fileMeta.Location = cephPath
	} else if cfg.CurrentStoreType == cmn.StoreOSS {
		// 文件写入OSS存储
		ossPath := "oss/" + fileMeta.FileName
		// 判断写入OSS为同步还是异步
		if !cfg.AsyncTransferEnable {
			err = oss.Bucket().PutObject(ossPath, newFile)
			if err != nil {
				log.Println("failed save to oss ,err:", err.Error())
				errCode = -5
				return
			}
			fileMeta.Location = ossPath
		} else {
			// 写入异步转移任务队列
			data := mq.TransferData{
				FileHash:      fileMeta.FileSha1,
				CurLocation:   fileMeta.Location,
				DestLocation:  ossPath,
				DestStoreType: cmn.StoreOSS,
			}
			pubData, _ := json.Marshal(data)
			pubSuc := mq.Publish(
				cfg.TransExchangeName,
				cfg.TransOSSRoutingKey,
				pubData,
			)
			if !pubSuc {
				// TODO: 当前发送转移信息失败，稍后重试
			}
		}
	}

	//6.  更新文件表记录
	_ = meta.UpdateFileMetaDB(fileMeta)

	username := c.Request.FormValue("username")
	// 更新用户文件表
	suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1,
		fileMeta.FileName, fileMeta.FileSize)
	if suc {
		c.Redirect(http.StatusFound, "/static/view/home.html")
	} else {
		errCode = -6
	}

}

// GetFileMetaHandler : 获取文件元信息
func GetFileMetaHandler(c *gin.Context) {

	filehash := c.Request.FormValue("filehash")
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"code": -2,
				"msg":  "GetFileMetaHandler error",
			},
		)
		return
	}

	if fMeta != nil {
		data, err := json.Marshal(fMeta)
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{
					"code": -3,
					"msg":  "GetFileMetaHandler error",
				},
			)
		}
		c.Data(http.StatusOK, "application/json", data)
	} else {
		c.JSON(http.StatusOK,
			gin.H{
				"code": -4,
				"msg":  "No such file",
			})
	}
}

// FileQueryHandler : 查询批量的文件元信息
func FileQueryHandler(c *gin.Context) {

	limitCnt, _ := strconv.Atoi(c.Request.FormValue("limit"))
	username := c.Request.FormValue("username")
	//fileMetas, _ := meta.GetLastFileMetasDB(limitCnt)
	userFiles, err := dblayer.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"code": -1,
				"msg":  "query failed",
			})
		return
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"code": -2,
				"msg":  "query failed",
			})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

// DownloadHandler : 文件下载接口
func DownloadHandler(c *gin.Context) {
	filehash := c.Request.FormValue("filehash")
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"code": -1,
				"msg":  "download err",
			})
		return
	}
	c.FileAttachment(fileMeta.Location, fileMeta.FileName)

}

// FileMetaUpdateHandler ： 更新元信息接口(重命名)
func FileMetaUpdateHandler(c *gin.Context) {

	opType := c.Request.FormValue("op")
	fileSha1 := c.Request.FormValue("filehash")
	username := c.Request.FormValue("username")
	newFileName := c.Request.FormValue("filename")

	if opType != "0" || len(newFileName) < 1 {
		c.JSON(http.StatusForbidden, gin.H{
			"code": -1,
			"msg":  "parameter error",
		})
		return
	}
	fileMeta, _ := meta.GetFileMetaDB(fileSha1)
	fileMeta.FileName = newFileName
	// 更新用户文件表tbl_user_file中的文件名，tbl_file的文件名不用修改
	meta.UpdateFileMeta(username, *fileMeta)

}

// FileDeleteHandler : 删除文件及元信息
func FileDeleteHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	fileSha1 := c.Request.FormValue("filehash")

	fm, err := meta.GetFileMetaDB(fileSha1)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"code": -1,
			"msg":  "file not found",
		})
		return
	}

	// 删除本地文件
	os.Remove(fm.Location)
	// TODO: 可考虑删除Ceph/OSS上的文件
	// 可以不立即删除，加个超时机制，
	// 比如该文件10天后也没有用户再次上传，那么就可以真正的删除了

	// 删除文件表中的一条记录
	suc := dblayer.DeleteUserFile(username, fileSha1)
	if !suc {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -2,
			"msg":  "failed delete file",
		})
		return
	}
	c.Status(http.StatusOK)
}

// TryFastUploadHandler : 尝试秒传接口
func TryFastUploadHandler(c *gin.Context) {
	// 1. 解析请求参数
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filename := c.Request.FormValue("filename")
	filesize, _ := strconv.Atoi(c.Request.FormValue("filesize"))

	// 2. 从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// 3. 查不到记录则返回秒传失败
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
		return
	}

	// 4. 上传过则将文件信息写入用户文件表， 返回成功
	suc := dblayer.OnUserFileUploadFinished(
		username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
		return
	}
	resp := util.RespMsg{
		Code: -2,
		Msg:  "秒传失败，请稍后重试",
	}
	c.Data(http.StatusOK, "application/json", resp.JSONBytes())
	return
}

// DownloadURLHandler : 生成文件的下载地址
func DownloadURLHandler(c *gin.Context) {
	filehash := c.Request.FormValue("filehash")
	// 从文件表查找记录
	row, _ := dblayer.GetFileMeta(filehash)

	// TODO: 判断文件存在OSS，还是Ceph，还是在本地
	if strings.HasPrefix(row.FileAddr.String, cfg.TempLocalRootDir) {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")
		// 返回下载链接
		tmpUrl := fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
			c.Request.Host, filehash, username, token)
		c.Data(http.StatusOK, "octet-stream", []byte(tmpUrl))
	} else if strings.HasPrefix(row.FileAddr.String, "/ceph") {
		// TODO: ceph下载url
	} else if strings.HasPrefix(row.FileAddr.String, "oss/") {
		// oss下载url
		signedURL := oss.DownloadURL(row.FileAddr.String)
		c.Data(http.StatusOK, "octet-stream", []byte(signedURL))
	}
}
