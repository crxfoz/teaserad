package tarantool

import (
	"context"
	"fmt"
	"testing"

	"github.com/crxfoz/teaserad/adshow/internal/domain/events"
	clustertnt "github.com/crxfoz/teaserad/adshow/pkg/tarantool"
	"github.com/stretchr/testify/assert"
	"github.com/tarantool/go-tarantool"
	"go.uber.org/zap"
)

func TestShowRepo_DeleteBannerAll(t *testing.T) {
	z, _ := zap.NewDevelopment()
	logger := z.Sugar()

	conn, err := tarantool.Connect("127.0.0.1:3301", tarantool.Opts{
		User: "admin",
		Pass: "pass",
	})
	assert.Nil(t, err)

	cluster := clustertnt.New([]*tarantool.Connection{conn})

	repo := New(cluster, logger)

	err = repo.DeleteBannerAll(context.Background(), 111)
	assert.Nil(t, err)
}

func TestShowRepo_AddBanner(t *testing.T) {
	z, _ := zap.NewDevelopment()
	logger := z.Sugar()

	conn, err := tarantool.Connect("127.0.0.1:3301", tarantool.Opts{
		User: "admin",
		Pass: "pass",
	})
	assert.Nil(t, err)

	cluster := clustertnt.New([]*tarantool.Connection{conn})

	repo := New(cluster, logger)

	banner := &events.BannerStart{
		BannerID:   111,
		UserID:     1231231,
		ImgData:    []byte("hello workd"),
		BannerText: "dasda",
		BannerURL:  "http://ya.ru",
		Device:     "desktop",
		CategoryID: 1,
	}

	err = repo.AddBanner(context.Background(), banner, []int{1, 2, 3})
	assert.Nil(t, err)
}

func TestShowRepo_BannersForPlatform(t *testing.T) {
	z, _ := zap.NewDevelopment()
	logger := z.Sugar()

	conn, err := tarantool.Connect("127.0.0.1:3301", tarantool.Opts{
		User: "admin",
		Pass: "pass",
	})
	assert.Nil(t, err)

	cluster := clustertnt.New([]*tarantool.Connection{conn})

	repo := New(cluster, logger)

	banners, err := repo.BannersForPlatform(context.Background(), 2, "desktop", 10)
	assert.Nil(t, err)

	for _, banner := range banners {
		fmt.Println(banner)
	}
}
