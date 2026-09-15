package service

import (
	"context"
	"log/slog"
	"os"
	"shortner/internal/domain"
	"shortner/internal/storage/types"
	"shortner/internal/tests"
	"shortner/internal/utils"
	storage_mock "shortner/mocks/storage"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type mockT struct {
	linkStorage *storage_mock.MockLinkStorage
}

func newMockT(t *testing.T) mockT {
	ctrl := gomock.NewController(t)
	return mockT{
		linkStorage: storage_mock.NewMockLinkStorage(ctrl),
	}
}

func newMockService(t *testing.T, mock mockT) *ShortnerService {
	return &ShortnerService{
		linkStorage: mock.linkStorage,
		logger:      slog.New(tests.NewTestHandler(t)),
	}
}

func TestService_ShortenLink(t *testing.T) {
	t.Parallel()

	mockURL := "https://ya.ru"
	mockAlias := "2Xy16t"

	type argsT struct {
		link string
	}

	tests := []struct {
		name          string
		args          argsT
		prepare       func(ctx context.Context, t *testing.T, args *argsT, mocks mockT)
		expectAlias   string
		expectErr     bool
		expectErrType error
	}{
		{
			name: "ok",
			args: argsT{
				link: mockURL,
			},
			prepare: func(_ context.Context, t *testing.T, args *argsT, mocks mockT) {
				input, err := domain.NewLink(args.link, utils.EncodeBase62(utils.GetHash(args.link)))
				require.NoError(t, err)

				mocks.linkStorage.EXPECT().PutLink(gomock.Any(), input).
					Return(input.Alias, nil)
			},
			expectAlias: "http://localhost:8080/" + mockAlias,
			expectErr:   false,
		},
		{
			name: "nok: internal storage error",
			args: argsT{
				link: mockURL,
			},
			prepare: func(_ context.Context, t *testing.T, args *argsT, mocks mockT) {
				input, err := domain.NewLink(args.link, utils.EncodeBase62(utils.GetHash(args.link)))
				require.NoError(t, err)

				mocks.linkStorage.EXPECT().PutLink(gomock.Any(), input).
					Return("", assert.AnError)
			},
			expectErr:     true,
			expectErrType: ErrNotSaved,
		},
		{
			name: "nok: shorten invalid link fail",
			args: argsT{
				link: "",
			},
			prepare: func(_ context.Context, t *testing.T, args *argsT, mocks mockT) {
				domain.NewLink(args.link, utils.EncodeBase62(utils.GetHash(args.link)))
			},
			expectErr:     true,
			expectErrType: ErrInvalidURL,
		},

		{
			name: "nok: alias already exists",
			args: argsT{
				link: mockURL,
			},
			prepare: func(_ context.Context, t *testing.T, args *argsT, mocks mockT) {
				input, err := domain.NewLink(args.link, utils.EncodeBase62(utils.GetHash(args.link)))
				require.NoError(t, err)

				mocks.linkStorage.EXPECT().PutLink(gomock.Any(), input).
					Return("", types.ErrAlreadyExists)
			},
			expectErr:     true,
			expectErrType: ErrAlreadyExists,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mocks := newMockT(t)
			s := newMockService(t, mocks)
			ctx := t.Context()

			tt.prepare(ctx, t, &tt.args, mocks)

			actualAlias, err := s.ShortenLink(ctx, tt.args.link)
			if tt.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErrType)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectAlias, actualAlias)
		})
	}
}

type MockStorage struct {
	data             map[string]*domain.Link
	alreadyExistsErr error
}

func NewMockStorage(exists bool) *MockStorage {
	mockStore := &MockStorage{data: make(map[string]*domain.Link)}
	if exists {
		mockStore.data["3XqGtZ"] = &domain.Link{Alias: "3XqGtZ",
			OriginalURL: "https://google.com"}
	}
	return mockStore
}

func (mockStorage *MockStorage) PutLink(ctx context.Context, link *domain.Link) (string, error) {
	if _, ok := mockStorage.data[link.Alias]; ok {
		return "", types.ErrAlreadyExists
	}
	mockStorage.data[link.Alias] = link
	return link.Alias, nil
}

func (mockStorage *MockStorage) GetLink(ctx context.Context, alias string) (*domain.Link, error) {
	if _, ok := mockStorage.data[alias]; !ok {
		return nil, types.ErrNotFound
	}
	return mockStorage.data[alias], nil
}

func TestShortnerService_GetOriginalLink(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	type fields struct {
		linkStorage types.LinkStorage
		logger      *slog.Logger
	}
	type args struct {
		ctx   context.Context
		alias string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *domain.Link
		wantErr error
	}{
		{
			name: "success",
			fields: fields{NewMockStorage(true),
				logger},
			args: args{ctx: context.Background(),
				alias: "3XqGtZ"},
			want: &domain.Link{Alias: "3XqGtZ",
				OriginalURL: "https://google.com"},
			wantErr: nil,
		},
		{
			name: "get not existing url fail",
			fields: fields{NewMockStorage(false),
				logger},
			args: args{ctx: context.Background(),
				alias: "3XqGtZ"},
			want:    nil,
			wantErr: ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ShortnerService{
				linkStorage: tt.fields.linkStorage,
				logger:      tt.fields.logger,
			}
			got, err := s.GetOriginalLink(tt.args.ctx, tt.args.alias)
			if err != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
