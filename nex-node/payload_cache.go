package nexnode

import (
	"debug/elf"
	"errors"
	"fmt"
	"os"
	"strings"

	controlapi "github.com/ConnectEverything/nex/control-api"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

type payloadCache struct {
	rootDir string
	log     *logrus.Logger
	nc      *nats.Conn
}

func NewPayloadCache(nc *nats.Conn, log *logrus.Logger, dir string) *payloadCache {
	return &payloadCache{
		rootDir: dir,
		log:     log,
		nc:      nc,
	}
}

func (c *payloadCache) GetPayloadFromBucket(request *controlapi.RunRequest) (*os.File, error) {
	// TODO - check for locally cached version

	bucket := request.Location.Host
	key := strings.Trim(request.Location.Path, "/")
	opts := []nats.JSOpt{}
	if len(strings.TrimSpace(request.JsDomain)) != 0 {
		opts = append(opts, nats.APIPrefix(request.JsDomain))
	}
	js, err := c.nc.JetStream(opts...)
	if err != nil {
		return nil, err
	}
	store, err := js.ObjectStore(bucket)
	if err != nil {
		return nil, err
	}
	filename := fmt.Sprintf("%s.%s", bucket, key)
	_, err = store.GetInfo(filename)
	if err != nil {
		return nil, err
	}
	// TODO: examine objInfo.Digest to get file hash to compare against locally hashed file
	err = store.GetFile(key, filename)
	if err != nil {
		return nil, err
	}
	c.log.WithField("filename", filename).WithField("bucket", bucket).WithField("key", key).Info("Downloaded workload bytes from bucket")

	elfFile, err := elf.Open(filename)
	if err != nil {
		c.log.WithError(err).Error("Failed to verify downloaded file is a static-linked elf binary")
	}
	defer elfFile.Close()
	err = verifyStatic(elfFile)
	if err != nil {
		c.log.WithError(err).Error("❌ Invalid ELF binary")
		return nil, err
	} else {
		c.log.Info("✅ Verified static-linked ELF binary")
	}

	return os.Open(filename)
}

func verifyStatic(elf *elf.File) error {
	for _, prog := range elf.Progs {
		if prog.ProgHeader.Type == 3 { // PT_INTERP
			return errors.New("elf binary contains at least one dynamically linked dependency")
		}
	}
	return nil
}