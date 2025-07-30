package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"avito_2023/internal/database"
	"avito_2023/internal/user/handler"
	"avito_2023/internal/user/model"
	"avito_2023/internal/user/repo/mocks"
)

type Suite struct {
	suite.Suite

	r       *gin.Engine
	repo    *mocks.RepoMock
	handler *handler.Handler
}

func (s *Suite) SetupSuite() {
	s.repo = &mocks.RepoMock{}
	s.handler = handler.NewHandler(s.repo)

	gin.SetMode(gin.TestMode)
	s.r = gin.Default()

	handler.Route(s.r, s.handler)
}

func TestSuite(t *testing.T) {
	suite.Run(t, &Suite{})
}

func (s *Suite) TestGetUserSegments() {
	testCases := []struct {
		name         string
		inputUserID  uint
		mockFc       func(ctx context.Context, userID uint) ([]*string, error)
		expectedCode int
		expectedResp string
	}{
		{
			name:        "get user segments",
			inputUserID: 1000,
			mockFc: func(ctx context.Context, userID uint) ([]*string, error) {
				slugs := [3]string{"test-slug-1", "test-slug-2", "test-slug-3"}
				var res []*string
				for i := 0; i < 3; i++ {
					res = append(res, &slugs[i])
				}
				return res, nil
			},
			expectedCode: http.StatusOK,
			expectedResp: `
				{
				  "user_id":  1000,
				  "segments": ["test-slug-1", "test-slug-2", "test-slug-3"]
				}
			`,
		},
		{
			name:        "segments not found",
			inputUserID: 1000,
			mockFc: func(ctx context.Context, userID uint) ([]*string, error) {
				return nil, database.ErrNotFound
			},
			expectedCode: http.StatusNotFound,
			expectedResp: `
				{
				  "error": "segments for user 1000 not found"
				}
			`,
		},
		{
			name:        "failed to get user segments",
			inputUserID: 1000,
			mockFc: func(ctx context.Context, userID uint) ([]*string, error) {
				return nil, fmt.Errorf("something went wrong")
			},
			expectedCode: http.StatusInternalServerError,
			expectedResp: `
				{
				  "error": "something went wrong"
				}
			`,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			if tc.mockFc != nil {
				s.repo.GetUserSegmentsFunc = tc.mockFc
			}

			res := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/user/%d", tc.inputUserID), nil)
			s.r.ServeHTTP(res, req)

			assert.Equal(t, tc.expectedCode, res.Code)

			if tc.expectedResp != "" {
				assert.JSONEq(t, tc.expectedResp, res.Body.String())
			}
		})
	}
}

func (s *Suite) TestGetUserHistory() {
	now := time.Now()

	testCases := []struct {
		name         string
		inputUserID  uint
		inputMonth   time.Month
		inputYear    int
		mockFc       func(ctx context.Context, userID uint, month uint, year uint) ([]*model.UserHistory, error)
		expectedCode int
		expectedResp string
	}{
		{
			name:        "get user history",
			inputUserID: 1000,
			inputMonth:  time.Now().Month(),
			inputYear:   time.Now().Year(),
			mockFc: func(ctx context.Context, userID uint, month uint, year uint) ([]*model.UserHistory, error) {
				userHistory := []model.UserHistory{
					{
						Slug:      "test-slug-1",
						CreatedAt: now,
					},
					{
						Slug:      "test-slug-2",
						CreatedAt: now.AddDate(0, -2, 0),
						DeletedAt: &now,
					},
					{
						Slug:      "test-slug-3",
						CreatedAt: now.AddDate(-1, 0, 0),
						DeletedAt: &now,
					},
				}
				history := make([]*model.UserHistory, len(userHistory))
				for i := 0; i < len(userHistory); i++ {
					history[i] = &userHistory[i]
				}
				return history, nil
			},
			expectedCode: http.StatusOK,
			expectedResp: fmt.Sprintf(`
				{
				  "user_id": 1000,
				  "history": [
				  	{
					  "slug": "test-slug-1",
					  "created_at": "%s",
					  "deleted_at": null
					},
				  	{
					  "slug": "test-slug-2",
					  "created_at": "%s",
					  "deleted_at": "%s"
					},
					{
					  "slug": "test-slug-3",
					  "created_at": "%s",
					  "deleted_at": "%s"
					}
				  ]
				}
			`, now.Format("2006-01-02T15:04:05.999999Z07:00"),
				now.AddDate(0, -2, 0).Format("2006-01-02T15:04:05.999999Z07:00"),
				now.Format("2006-01-02T15:04:05.999999Z07:00"),
				now.AddDate(-1, 0, 0).Format("2006-01-02T15:04:05.999999Z07:00"),
				now.Format("2006-01-02T15:04:05.999999Z07:00")),
		},
		{
			name:         "invalid request uri (user_id)",
			expectedCode: http.StatusBadRequest,
			expectedResp: "{\"error\":\"Key: 'GetUserHistoryUri.UserID' Error:Field validation for 'UserID' failed on the 'required' tag\"}",
		},
		{
			name:         "invalid request query (month)",
			inputUserID:  1000,
			expectedCode: http.StatusBadRequest,
			expectedResp: "{\"error\":\"Key: 'GetUserHistoryQuery.Month' Error:Field validation for 'Month' failed on the 'required' tag\\nKey: 'GetUserHistoryQuery.Year' Error:Field validation for 'Year' failed on the 'required' tag\"}",
		},
		{
			name:        "history not found",
			inputUserID: 1000,
			inputMonth:  time.Now().Month(),
			inputYear:   time.Now().Year(),
			mockFc: func(ctx context.Context, userID uint, month uint, year uint) ([]*model.UserHistory, error) {
				return nil, database.ErrNotFound
			},
			expectedCode: http.StatusNotFound,
			expectedResp: `
				{
				  "error": "history for user 1000 not found"
				}
			`,
		},
		{
			name:        "failed to get history from db",
			inputUserID: 1000,
			inputMonth:  time.Now().Month(),
			inputYear:   time.Now().Year(),
			mockFc: func(ctx context.Context, userID uint, month uint, year uint) ([]*model.UserHistory, error) {
				return nil, fmt.Errorf("something went wrong")
			},
			expectedCode: http.StatusInternalServerError,
			expectedResp: `
				{
				  "error": "something went wrong"
				}
			`,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			if tc.mockFc != nil {
				s.repo.GetUserHistoryFunc = tc.mockFc
			}

			res := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/user/history/%d?month=%d&year=%d", tc.inputUserID, tc.inputMonth, tc.inputYear), nil)
			s.r.ServeHTTP(res, req)

			assert.Equal(t, tc.expectedCode, res.Code)
			t.Log(res.Body.String())

			if tc.expectedResp != "" {
				assert.JSONEq(t, tc.expectedResp, res.Body.String())
			}
		})
	}
}

func (s *Suite) TestUpdateUserSegments() {
	testCases := []struct {
		name         string
		inputBody    map[string]interface{}
		mockFc       func(ctx context.Context, userID uint, slugsToAdd, slugsToDel []string, deleteAt *time.Time) error
		expectedCode int
		expectedResp string
	}{
		{
			name: "update user segments",
			inputBody: map[string]interface{}{
				"user_id":      1000,
				"slugs_to_add": []string{"test-slug-1", "test-slug-2", "test-slug-3"},
				"slugs_to_del": []string{"test-slug-4"},
			},
			mockFc: func(ctx context.Context, userID uint, slugsToAdd, slugsToDel []string, deleteAt *time.Time) error {
				return nil
			},
			expectedCode: http.StatusNoContent,
		},
		{
			name: "update user segments",
			inputBody: map[string]interface{}{
				"user_id":      1000,
				"slugs_to_add": []string{"test-slug-1", "test-slug-2", "test-slug-3"},
				"slugs_to_del": []string{"test-slug-4"},
				"delete_at":    time.Now().AddDate(0, 0, 2).Unix(),
			},
			mockFc: func(ctx context.Context, userID uint, slugsToAdd, slugsToDel []string, deleteAt *time.Time) error {
				return nil
			},
			expectedCode: http.StatusNoContent,
		},
		{
			name: "invalid request body",
			inputBody: map[string]interface{}{
				"wrong": "wrong",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid request body (slugs_to_del)",
			inputBody: map[string]interface{}{
				"user_id":      1000,
				"slugs_to_add": []string{"test-slug-1", "test-slug-2", "test-slug-3"},
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid request body (slugs_to_add)",
			inputBody: map[string]interface{}{
				"user_id":      1000,
				"slugs_to_del": []string{"test-slug-1"},
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid request body (delete_at)",
			inputBody: map[string]interface{}{
				"user_id":      1000,
				"slugs_to_add": []string{"test-slug-1", "test-slug-2", "test-slug-3"},
				"slugs_to_del": []string{"test-slug-4"},
				"delete_at":    time.Now().AddDate(0, 0, -2).Unix(),
			},
			mockFc: func(ctx context.Context, userID uint, slugsToAdd, slugsToDel []string, deleteAt *time.Time) error {
				return nil
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid segments",
			inputBody: map[string]interface{}{
				"user_id":      1000,
				"slugs_to_add": []string{"wrong-slug-1", "wrong-slug-2", "wrong-slug-3"},
				"slugs_to_del": []string{"test-slug-1"},
			},
			mockFc: func(ctx context.Context, userID uint, slugsToAdd, slugsToDel []string, deleteAt *time.Time) error {
				return database.ErrUpdateUserSegments_InvalidSegments
			},
			expectedCode: http.StatusBadRequest,
			expectedResp: `
				{
				  "error": "invalid segments"
				}
			`,
		},
		{
			name: "failed to add user segments to db",
			inputBody: map[string]interface{}{
				"user_id":      1000,
				"slugs_to_add": []string{"test-slug-1", "test-slug-2", "test-slug-3"},
				"slugs_to_del": []string{"test-slug-4"},
			},
			mockFc: func(ctx context.Context, userID uint, slugsToAdd, slugsToDel []string, deleteAt *time.Time) error {
				return fmt.Errorf("something went wrong")
			},
			expectedCode: http.StatusInternalServerError,
			expectedResp: `
				{
				  "error": "something went wrong"
				}
			`,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			if tc.mockFc != nil {
				s.repo.UpdateUserSegmentsFunc = tc.mockFc
			}

			t.Log(time.Now().Unix())

			b, _ := json.Marshal(tc.inputBody)
			res := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPut, "/user/segment", bytes.NewBuffer(b))
			s.r.ServeHTTP(res, req)

			assert.Equal(t, tc.expectedCode, res.Code)

			if tc.expectedResp != "" {
				assert.JSONEq(t, tc.expectedResp, res.Body.String())
			}
		})
	}
}
