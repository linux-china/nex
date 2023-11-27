package nexnode

import (
	"io"
	"os"
	"path"
	"strings"

	agentapi "github.com/ConnectEverything/nex/agent-api"
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

func (m *MachineManager) CacheWorkload(request *controlapi.RunRequest) error {
	bucket := request.Location.Host
	key := strings.Trim(request.Location.Path, "/")
	m.log.WithField("bucket", bucket).WithField("key", key).WithField("url", m.nc.Opts.Url).Info("Attempting object store download")
	opts := []nats.JSOpt{}
	if len(strings.TrimSpace(request.JsDomain)) != 0 {
		opts = append(opts, nats.APIPrefix(request.JsDomain))
	}
	js, err := m.nc.JetStream(opts...)
	if err != nil {
		return err
	}
	store, err := js.ObjectStore(bucket)
	if err != nil {
		m.log.WithError(err).WithField("bucket", bucket).Error("Failed to bind to source object store")
		return err
	}
	_, err = store.GetInfo(key)
	if err != nil {
		m.log.WithError(err).WithField("key", key).WithField("bucket", bucket).Error("Failed to locate workload binary in source object store")
		return err
	}

	// NOTE: it's not the best use of time to write the file to disk and then read it back again in order to
	// hand it to the cache.. but the elf verification library only works on a filename and doesn't take
	// a reader or a slice

	filename := path.Join(os.TempDir(), "sus")
	err = store.GetFile(key, filename)
	if err != nil {
		m.log.WithError(err).WithField("key", key).Error("Failed to download bytes from source object store")
		return err
	}

	// TODO: as we support more workload types, validate those accordingly. For now, the node only
	// supports 64-bit statically linked binaries
	err = controlapi.ValidateNativeBinary(filename)
	if err != nil {
		m.log.WithError(err).Error("❌ Invalid ELF binary")
		return err
	} else {
		m.log.WithField("key", key).WithField("name", request.DecodedClaims.Subject).Info("✅ Verified static-linked ELF binary")
	}

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	workload, err := io.ReadAll(f)
	if err != nil {
		m.log.WithError(err).Error("Couldn't read the file we just wrote")
		return err
	}
	os.Remove(filename)

	jsInternal, err := m.ncInternal.JetStream()
	if err != nil {
		m.log.WithError(err).Error("Failed to acquire JetStream context for internal object store.")
		panic(err)
	}
	cache, err := jsInternal.ObjectStore(agentapi.WorkloadCacheBucket)
	if err != nil {
		m.log.WithError(err).Error("Failed to get object store reference for internal cache.")
		panic(err)
	}

	_, err = cache.PutBytes(request.DecodedClaims.Subject, workload)
	if err != nil {
		m.log.WithError(err).Error("Failed to write workload to internal cache.")
		panic(err)
	}

	m.log.WithField("name", request.DecodedClaims.Subject).Info("Successfully stored workload in internal object store")

	return nil
}
