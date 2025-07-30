package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"avito_2023/internal/database"
	"avito_2023/internal/segment/handler"
	"avito_2023/internal/segment/repo/mocks"
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

func (s *Suite) TestAddSegment() {
	testCases := []struct {
		name         string
		inputBody    map[string]interface{}
		mockFc       func(ctx context.Context, slug string, percentage uint) error
		expectedCode int
		expectedErr  string
	}{
		{
			name: "add segment",
			inputBody: map[string]interface{}{
				"slug": "test-slug",
			},
			mockFc: func(ctx context.Context, slug string, percentage uint) error {
				return nil
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "add segment with percentage",
			inputBody: map[string]interface{}{
				"slug":       "test-slug",
				"percentage": 10,
			},
			mockFc: func(ctx context.Context, slug string, percentage uint) error {
				return nil
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "invalid request body",
			inputBody: map[string]interface{}{
				"wrong": "wrong",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid request body (percentage)",
			inputBody: map[string]interface{}{
				"slug":       "test-slug",
				"percentage": 101,
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "failed to add segment to db",
			inputBody: map[string]interface{}{
				"slug": "test-slug",
			},
			mockFc: func(ctx context.Context, slug string, percentage uint) error {
				return fmt.Errorf("something went wrong")
			},
			expectedCode: http.StatusInternalServerError,
			expectedErr:  "something went wrong",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			if tc.mockFc != nil {
				s.repo.AddSegmentFunc = tc.mockFc
			}

			b, _ := json.Marshal(tc.inputBody)
			res := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/segment/add", bytes.NewBuffer(b))
			s.r.ServeHTTP(res, req)

			assert.Equal(s.T(), tc.expectedCode, res.Code)

			if tc.expectedErr != "" {
				assert.Contains(t, res.Body.String(), tc.expectedErr)
			}
		})
	}
}

func (s *Suite) TestDeleteSegment() {
	testCases := []struct {
		name         string
		inputBody    map[string]interface{}
		mockFc       func(ctx context.Context, slug string) error
		expectedCode int
		expectedErr  string
	}{
		{
			name: "delete segment",
			inputBody: map[string]interface{}{
				"slug": "test-slug",
			},
			mockFc: func(ctx context.Context, slug string) error {
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
			name: "segment not found",
			inputBody: map[string]interface{}{
				"slug": "test-slug",
			},
			mockFc: func(ctx context.Context, slug string) error {
				return database.ErrNotFound
			},
			expectedCode: http.StatusNotFound,
			expectedErr:  "segment test-slug not found",
		},
		{
			name: "failed to delete segment from db",
			inputBody: map[string]interface{}{
				"slug": "test-slug",
			},
			mockFc: func(ctx context.Context, slug string) error {
				return fmt.Errorf("something went wrong")
			},
			expectedCode: http.StatusInternalServerError,
			expectedErr:  "something went wrong",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			if tc.mockFc != nil {
				s.repo.DeleteSegmentFunc = tc.mockFc
			}

			b, _ := json.Marshal(tc.inputBody)
			res := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodDelete, "/segment/delete", bytes.NewBuffer(b))
			s.r.ServeHTTP(res, req)

			assert.Equal(t, tc.expectedCode, res.Code)

			if tc.expectedErr != "" {
				assert.Contains(t, res.Body.String(), tc.expectedErr)
			}
		})
	}
}
