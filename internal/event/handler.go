package event

import (
	"event-service/helper"
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

	events, err := h.eventService.GetAllEvents(c)
	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, err, helper.ErrInvalidOperation)
		return
	}

	helper.SendSuccess(c, http.StatusOK, "success", events)

}
