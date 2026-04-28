package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const baseURL = "http://localhost:8080"

func getAuthToken(t *testing.T, role string) string {
	reqBody, _ := json.Marshal(map[string]string{
		"role": role,
	})
	resp, err := http.Post(baseURL+"/dummyLogin", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("dummyLogin failed with status %d", resp.StatusCode)
	}

	var res map[string]any
	json.NewDecoder(resp.Body).Decode(&res)

	token, ok := res["token"].(string)
	if !ok {
		t.Fatalf("No token in response")
	}
	return token
}

func doRequest(t *testing.T, method, path, token string, body any) (*http.Response, map[string]any) {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req, err := http.NewRequest(method, baseURL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute request %s %s: %v", method, path, err)
	}

	var res map[string]any
	if resp.StatusCode != http.StatusNoContent {
		json.NewDecoder(resp.Body).Decode(&res)
		resp.Body.Close()
	}
	return resp, res
}

func TestE2E_CreateRoom_CreateSchedule_CreateBooking(t *testing.T) {
	resp, err := http.Get(baseURL + "/_info")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Skip("E2E tests skipped: Server not reachable at " + baseURL)
	}

	adminToken := getAuthToken(t, "admin")
	userToken := getAuthToken(t, "user")

	roomName := fmt.Sprintf("E2E Room %d", time.Now().UnixNano())

	roomResp, roomRes := doRequest(t, "POST", "/rooms/create", adminToken, map[string]any{
		"name":        roomName,
		"description": "Integration test room",
		"capacity":    10,
	})
	if roomResp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create room: status %d, body: %v", roomResp.StatusCode, roomRes)
	}
	roomId, ok := roomRes["room"].(map[string]any)["id"].(string)
	if !ok || roomId == "" {
		t.Fatalf("Invalid roomId returned")
	}

	tNow := time.Now().UTC()
	startStr := tNow.Add(-1 * time.Hour).Format("15:04")
	endStr := tNow.Add(2 * time.Hour).Format("15:04")

	scheduleResp, scheduleRes := doRequest(t, "POST", "/rooms/"+roomId+"/schedule/create", adminToken, map[string]any{
		"daysOfWeek": []int{1, 2, 3, 4, 5, 6, 7},
		"startTime":  "00:00",
		"endTime":    "23:59",
	})
	if scheduleResp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create schedule: status %d, expected 201. body: %v, start: %s, end: %s", scheduleResp.StatusCode, scheduleRes, startStr, endStr)
	}

	dateStr := tNow.AddDate(0, 0, 1).Format("2006-01-02")
	slotsResp, slotsRes := doRequest(t, "GET", "/rooms/"+roomId+"/slots/list?date="+dateStr, userToken, nil)
	if slotsResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to list slots: status %d", slotsResp.StatusCode)
	}

	slots, ok := slotsRes["slots"].([]any)
	if !ok || len(slots) == 0 {
		t.Fatalf("No slots found for created schedule.")
	}
	firstSlot := slots[0].(map[string]any)
	slotId := firstSlot["id"].(string)

	bookingResp, bookingRes := doRequest(t, "POST", "/bookings/create", userToken, map[string]any{
		"slotId": slotId,
	})
	if bookingResp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create booking: status %d, body: %v", bookingResp.StatusCode, bookingRes)
	}

	bookingId, ok := bookingRes["booking"].(map[string]any)["id"].(string)
	if !ok {
		t.Fatalf("Booking wasn't properly created")
	}
	t.Logf("Successfully created booking %s", bookingId)
}

func TestE2E_CancelBooking(t *testing.T) {

	resp, err := http.Get(baseURL + "/_info")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Skip("E2E tests skipped: Server not reachable at " + baseURL)
	}

	adminToken := getAuthToken(t, "admin")
	userToken := getAuthToken(t, "user")

	roomName := fmt.Sprintf("E2E Cancel Room %d", time.Now().UnixNano())

	_, roomRes := doRequest(t, "POST", "/rooms/create", adminToken, map[string]any{
		"name":        roomName,
		"description": "Integration test room for cancel",
		"capacity":    5,
	})
	roomId := roomRes["room"].(map[string]any)["id"].(string)

	doRequest(t, "POST", "/rooms/"+roomId+"/schedule/create", adminToken, map[string]any{
		"daysOfWeek": []int{1, 2, 3, 4, 5, 6, 7},
		"startTime":  "00:00",
		"endTime":    "23:59",
	})

	dateStr := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	_, slotsRes := doRequest(t, "GET", "/rooms/"+roomId+"/slots/list?date="+dateStr, userToken, nil)
	slots := slotsRes["slots"].([]any)
	if len(slots) == 0 {
		t.Fatalf("No slots available to test cancel")
	}
	slotId := slots[1].(map[string]any)["id"].(string)

	_, bookingRes := doRequest(t, "POST", "/bookings/create", userToken, map[string]any{
		"slotId": slotId,
	})
	bookingId := bookingRes["booking"].(map[string]any)["id"].(string)

	cancelResp, cancelRes := doRequest(t, "POST", "/bookings/"+bookingId+"/cancel", userToken, nil)
	if cancelResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to cancel booking: status: %d, body: %v", cancelResp.StatusCode, cancelRes)
	}

	status := cancelRes["booking"].(map[string]any)["status"].(string)
	if status != "cancelled" {
		t.Fatalf("Expected status to be cancelled, got %s", status)
	}
	t.Logf("Successfully cancelled booking %s", bookingId)
}
