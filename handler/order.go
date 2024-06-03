package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tomerlevy1/go-orders-api/model"
	"github.com/tomerlevy1/go-orders-api/repository/order"
)

type Order struct {
	Repo *order.RedisRepo
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID uuid.UUID        `json:"customer_id"`
		LineItems  []model.LineItem `json:"line_items"`
	}

	if err := json.NewDecoder(r.Body).Decode((&body)); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	order := model.Order{
		OrderID:    rand.Uint64(),
		CustomerID: body.CustomerID,
		LineItems:  body.LineItems,
		CreatedAt:  &now,
	}

	err := o.Repo.Insert(r.Context(), order)
	if err != nil {
		fmt.Errorf("failed to insert order: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(order)
	if err != nil {
		fmt.Errorf("failed to marshal order: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(res)
	w.WriteHeader(http.StatusCreated)
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	cursor := q.Get("cursor")
	if cursor == "" {
		cursor = "0"
	}

	res, err := strconv.ParseUint(cursor, 10, 64)
	if err != nil {
		fmt.Errorf("failed to parse cursor: %w", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	const size = 50
	orders, err := o.Repo.FindAll(r.Context(), order.FindAllPage{
		Offset: res,
		Size:   size,
	})

	if err != nil {
		fmt.Errorf("failed to find all orders: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response struct {
		Items []model.Order `json:"items"`
		Next  uint64        `json:"next,omitempty"`
	}

	response.Items = orders.Orders
	response.Next = orders.Cursor

	result, err := json.Marshal(response)
	if err != nil {
		fmt.Errorf("failed to marshal response: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(result)
}

func (h *Order) GetById(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	o, err := h.Repo.FindByID(r.Context(), id)
	if errors.Is(err, order.ErrorNotExists) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, err := json.Marshal(o)
	if err != nil {
		fmt.Errorf("failed to marshal order: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(result)
}

func (h *Order) UpdateById(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		fmt.Println("failed to decode body:", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request body"))
		return
	}

	idParam := chi.URLParam(r, "id")

	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid order ID"))
		return
	}

	o, err := h.Repo.FindByID(r.Context(), id)
	if errors.Is(err, order.ErrorNotExists) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Order not found"))
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	const completedStatus = "completed"
	const shippedStatus = "shipped"
	now := time.Now().UTC()

	switch body.Status {
	case completedStatus:
		fmt.Println(o.CompletedAt)
		if o.ShippedAt == nil || o.CompletedAt != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Order already completed or not shipped yet"))
			return
		}
		o.CompletedAt = &now
	case shippedStatus:
		if o.ShippedAt != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Order already shipped"))
			return
		}
		o.ShippedAt = &now
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid status"))
		return
	}

	err = h.Repo.UpdateByID(r.Context(), o)
	if err != nil {
		fmt.Println("failed to update order: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	if err := json.NewEncoder(w).Encode(o); err != nil {
		fmt.Println("failed to marshal order: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}
}

func (o *Order) DeleteById(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid order ID"))
		return
	}

	err = o.Repo.DeleteByID(r.Context(), id)
	if errors.Is(err, order.ErrorNotExists) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Order not found"))
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
