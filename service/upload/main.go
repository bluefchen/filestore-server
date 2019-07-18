package main

import (
	cfg "filestore-server/config"
	"filestore-server/route"
)

func main() {
	/*// 静态资源处理
	// http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(assets.AssetFS())))
	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("../../static"))))

		handler.TryFastUploadHandler))
	// 获取文件下载地址
	http.HandleFunc("/file/downloadurl", handler.HTTPInterceptor(
		handler.DownloadURLHandler))

	// 分块上传接口
	//初始化上传信息
	http.HandleFunc("/file/mpupload/init",
		handler.HTTPInterceptor(handler.InitialMultipartUploadHandler))
	//分块上传
	http.HandleFunc("/file/mpupload/uppart",
		handler.HTTPInterceptor(handler.UploadPartHandler))
	//上传完成：合并、记录数据
	http.HandleFunc("/file/mpupload/complete",
		handler.HTTPInterceptor(handler.CompleteUploadHandler))

	// 用户相关接口*/

	router := route.Router()
	router.Run(cfg.UploadServiceHost)

}
