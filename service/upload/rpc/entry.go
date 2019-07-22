package rpc

import (
	"context"
	"filestore-server/service/upload/config"
	proto "filestore-server/service/upload/proto"
)

type Upload struct {

}

func (this *Upload) UploadEntry(ctx context.Context, req *proto.ReqEntry, resp *proto.RespEntry) error {
	resp.Entry=config.UploadEntry
	return nil
}
