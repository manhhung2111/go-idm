package grpc

import (
	"context"
	"fmt"

	go_idm_v1 "github.com/manhhung2111/go-idm/internal/generated/proto"
	"github.com/manhhung2111/go-idm/internal/logic"
)

type Handler struct {
	go_idm_v1.UnimplementedGoIDMServiceServer
	accountLogic logic.Account
	downloadTaskLogic logic.DownloadTask
}

func NewHandler(
	accountLogic logic.Account,
	downloadTaskLogic logic.DownloadTask,
) go_idm_v1.GoIDMServiceServer {
	return &Handler{
		accountLogic: accountLogic,
		downloadTaskLogic: downloadTaskLogic,
	}
}

func (h *Handler) CreateAccount(ctx context.Context, req *go_idm_v1.CreateAccountRequest) (*go_idm_v1.CreateAccountResponse, error) {
	output, err := h.accountLogic.CreateAccount(ctx, logic.CreateAccountParams{
		AccountName: req.GetAccountName(),
		Password:    req.GetPassword(),
	})

	if err != nil {
		return nil, err
	}

	return &go_idm_v1.CreateAccountResponse{
		AccountId: output.ID,
	}, nil
}

func (h *Handler) CreateSession(ctx context.Context, req *go_idm_v1.CreateSessionRequest) (*go_idm_v1.CreateSessionResponse, error) {
	token, err := h.accountLogic.CreateSession(ctx, logic.CreateSessionParams{
		AccountName: req.GetAccountName(),
		Password:    req.GetPassword(),
	})
	if err != nil {
		return nil, err
	}

	return &go_idm_v1.CreateSessionResponse{
		Token: token,
	}, nil
}

func (h *Handler) CreateDownloadTask(ctx context.Context, req *go_idm_v1.CreateDownloadTaskRequest) (*go_idm_v1.CreateDownloadTaskResponse, error) {
	output, err := h.downloadTaskLogic.CreateDownloadTask(ctx, logic.CreateDownloadTaskParams{
		Token:        req.GetToken(),
		DownloadType: req.GetDownloadType(),
		URL:          req.GetUrl(),
	})
	if err != nil {
		return nil, err
	}

	return &go_idm_v1.CreateDownloadTaskResponse{
		DownloadTask: &output.DownloadTask,
	}, nil
}

func (h *Handler) GetDownloadTaskList(ctx context.Context, req *go_idm_v1.GetDownloadTaskListRequest) (*go_idm_v1.GetDownloadTaskListResponse, error) {
	fmt.Println("GetDownloadTaskList called")
	return &go_idm_v1.GetDownloadTaskListResponse{}, nil
}

func (h *Handler) UpdateDownloadTask(ctx context.Context, req *go_idm_v1.UpdateDownloadTaskRequest) (*go_idm_v1.UpdateDownloadTaskResponse, error) {
	fmt.Println("UpdateDownloadTask called")
	return &go_idm_v1.UpdateDownloadTaskResponse{}, nil
}

func (h *Handler) DeleteDownloadTask(ctx context.Context, req *go_idm_v1.DeleteDownloadTaskRequest) (*go_idm_v1.DeleteDownloadTaskResponse, error) {
	fmt.Println("DeleteDownloadTask called")
	return &go_idm_v1.DeleteDownloadTaskResponse{}, nil
}

func (h *Handler) GetDownloadTaskFile(req *go_idm_v1.GetDownloadTaskFiletRequest, stream go_idm_v1.GoIDMService_GetDownloadTaskFileServer) error {
	fmt.Println("GetDownloadTaskFile called")
	resp := &go_idm_v1.GetDownloadTaskFiletResponse{}
	if err := stream.Send(resp); err != nil {
		return err
	}
	return nil
}
