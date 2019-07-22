package api

import (
	"bytes"
	"encoding/json"
	"filestore-server/meta"
	"filestore-server/mq"
	"filestore-server/store/ceph"
	"filestore-server/store/oss"
	"filestore-server/util"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	cfg "filestore-server/config"
	cmn "filestore-server/common"
	dblayer "filestore-server/db"
)

func UploadHandler(c *gin.Context) {
	data, err := ioutil.ReadFile("../../static/view/upload.html")
	if err != nil {
		c.String(404, `网页不存在`)
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, string(data))
}

//DoUploadHandler：处理上传文件
func DoUploadHandler(c *gin.Context) {
	errCode := 0
	//返回前调用
	defer func() {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST,OPTIONS")
		if errCode < 0 {
			c.JSON(http.StatusOK, gin.H{
				"code": errCode,
				"msg":  "Upload Failed",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": errCode,
				"msg":  "上传成功",
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
		errCode = 0
	} else {
		errCode = -6
	}

}
