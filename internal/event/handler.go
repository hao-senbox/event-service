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

	helper.SendSuccess(c, http.StatusCreated, "created event successfully", nil)

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

	helper.SendSuccess(c, http.StatusOK, "Get all events successfully", events)

}

func (h *EventHandler) GetEventByID(c *gin.Context) {

	id := c.Param("id")

	event, err := h.eventService.GetEventByID(c, id)
	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, err, helper.ErrInvalidOperation)
		return
	}

	helper.SendSuccess(c, http.StatusOK, "Get event successfully", event)

}

func (h *EventHandler) UpdateEvent(c *gin.Context) {

	id := c.Param("id")

	if id == "" {
		helper.SendError(c, http.StatusBadRequest, fmt.Errorf("id is required"), helper.ErrInvalidRequest)
		return
	}

	var req UpdateEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendError(c, http.StatusBadRequest, err, helper.ErrInvalidRequest)
		return
	}

	err := h.eventService.UpdateEvent(c, &req, id)
	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, err, helper.ErrInvalidOperation)
		return
	}

	helper.SendSuccess(c, http.StatusOK, "Update event successfully", nil)
}

func (h *EventHandler) DeleteEvent(c *gin.Context) {

	id := c.Param("id")

	err := h.eventService.DeleteEvent(c, id)
	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, err, helper.ErrInvalidOperation)
		return
	}

	helper.SendSuccess(c, http.StatusOK, "Delete event successfully", nil)
}

func (h *EventHandler) ToggleSendEventNotifications(c *gin.Context) {

	id := c.Param("id")

	check, err := h.eventService.ToggleSendEventNotifications(c, id)
	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, err, helper.ErrInvalidOperation)
		return
	}

	helper.SendSuccess(c, http.StatusOK, "Change status successfully", check)

}

func (h *EventHandler) ToggleShowEventNotifications(c *gin.Context) {

	id := c.Param("id")

	check, err := h.eventService.ToggleShowEventNotifications(c, id)
	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, err, helper.ErrInvalidOperation)
		return
	}

	helper.SendSuccess(c, http.StatusOK, "Change status successfully", check)

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

	helper.SendSuccess(c, http.StatusOK, "Send event notifications successfully", nil)
}
