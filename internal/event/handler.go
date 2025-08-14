package event

import (
	"context"
	"event-service/helper"
	"event-service/pkg/constants"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	eventService EventService
}

func NewEventHandler(eventService EventService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

func (h *EventHandler) CreateEvent(c *gin.Context) {

	var req CreateEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendError(c, http.StatusBadRequest, err, helper.ErrInvalidRequest)
		return
	}

	err := h.eventService.CreateEvent(c, &req)
	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, err, helper.ErrInvalidOperation)
		return
	}

	helper.SendSuccess(c, http.StatusCreated, "success", nil)

}

func (h *EventHandler) GetAllEvents(c *gin.Context) {

	userID := c.Query("user_id")
	if userID == "" {
		helper.SendError(c, http.StatusBadRequest, fmt.Errorf("user_id is required"), helper.ErrInvalidRequest)
		return
	}

	events, err := h.eventService.GetAllEvents(c, userID)
	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, err, helper.ErrInvalidOperation)
		return
	}

	helper.SendSuccess(c, http.StatusOK, "success", events)

}

func (h *EventHandler) SendEventNotifications(c *gin.Context) {

	var req TriggerEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendError(c, http.StatusBadRequest, err, helper.ErrInvalidRequest)
		return
	}

	token, exists := c.Get(constants.Token)
	if !exists {
		helper.SendError(c, 400, fmt.Errorf("token not found"), helper.ErrInvalidRequest)
		return
	}
	
	ctx := context.WithValue(c, constants.TokenKey, token)

	err := h.eventService.SendEventNotifications(ctx, &req)
	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, err, helper.ErrInvalidOperation)
		return
	}

	helper.SendSuccess(c, http.StatusOK, "success", nil)
}