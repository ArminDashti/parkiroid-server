package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/dogan/dogan-server/internal/models"
	"github.com/dogan/dogan-server/internal/store"
	"github.com/gin-gonic/gin"
)

type ActionsHandler struct {
	actionStore store.ActionStore
}

func NewActionsHandler(actionStore store.ActionStore) *ActionsHandler {
	return &ActionsHandler{actionStore: actionStore}
}

func (handler *ActionsHandler) CreateAction(context *gin.Context) {
	var payload models.PhoneActionPayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	action := models.PhoneActionRecord{
		DeviceID:   payload.DeviceID,
		ActionType: payload.ActionType,
		Payload:    payload.Payload,
		SentAt:     time.Now().UTC(),
		Status:     "pending",
	}

	createdAction, err := handler.actionStore.CreateAction(action)
	if err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to create action"})
		return
	}

	context.JSON(http.StatusCreated, createdAction)
}

func (handler *ActionsHandler) GetPendingActions(context *gin.Context) {
	deviceID := context.Query("device-id")
	if deviceID == "" {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "device-id query parameter is required"})
		return
	}

	actions, err := handler.actionStore.GetPendingActions(deviceID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to retrieve pending actions"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"actions": actions})
}

func (handler *ActionsHandler) AcknowledgeAction(context *gin.Context) {
	actionIDRaw := context.Param("id")
	actionID, err := strconv.ParseInt(actionIDRaw, 10, 64)
	if err != nil || actionID <= 0 {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid action id"})
		return
	}

	var payload models.PhoneActionAckPayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	status := payload.Status
	if status == "" {
		status = "done"
	}

	if err := handler.actionStore.AcknowledgeAction(actionID, status); err != nil {
		if errors.Is(err, store.ErrActionNotFound) {
			context.JSON(http.StatusNotFound, models.ErrorResponse{Error: "action not found"})
			return
		}
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to acknowledge action"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"id": actionID, "status": status})
}
