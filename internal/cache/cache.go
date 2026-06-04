package cache

import "github.com/Bupyc/link-storage-service/internal/model"

type LinkCache interface {
	Get(shortCode string) (model.Link, bool)
	Set(link model.Link)
	Delete(shortCode string)
}
