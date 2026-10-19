package infra

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ddd-example/internal/migrate"
	"ddd-example/pkg/database"

	"github.com/jmoiron/sqlx"
	"github.com/joyparty/entity"
	"github.com/joyparty/entity/cache"

	// database driver
	_ "github.com/mattn/go-sqlite3"
)

// specTimeout 单个 spec 的超时时间（Ginkgo SpecTimeout 装饰器）
//
// 防止 spec 内的数据库操作异常挂起（例如等待锁），让失败快速暴露。
const specTimeout = SpecTimeout(10 * time.Second)

var testDB *sqlx.DB

func TestInfra(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Infra Suite")
}

// testDSN 基于数据库文件路径构造 DSN
//
// 参数与生产 DSN 保持一致（internal/bootstrap/bootstrap.go 的 getDBDSN）：
// WAL、_sync=NORMAL、_fk=1、busy timeout 等，保证测试与生产行为一致。
// _timeout 与驱动默认的 busy_timeout 相同（5s），显式写出以对齐生产配置。
//
// 额外的 _txlock=immediate：多进程共享同一个 SQLite 文件时，事务在 BEGIN
// 阶段就取得写锁（busy timeout 内会等待而不是立即失败），避免"先读后写"的
// 锁升级在多进程下直接返回 SQLITE_BUSY(database is locked)。
func testDSN(dbPath string) string {
	return dbPath + "?mode=rwc&_timeout=5000&_journal=WAL&_sync=NORMAL&_encoding=UTF-8&_fk=1&_txlock=immediate"
}

// 所有进程共享同一个文件数据库：进程 1 创建临时目录、建库并执行 migrate，
// 数据库文件路径通过 SynchronizedBeforeSuite 广播给所有进程；
// 进程 1 注册的 cleanup 会等其余进程全部结束后才执行，届时删除临时目录。
var _ = SynchronizedBeforeSuite(func() []byte {
	tmpDir, err := os.MkdirTemp("", "ddd-example-infra-*")
	Expect(err).NotTo(HaveOccurred(), "create test db dir")
	DeferCleanup(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed(), "remove test db dir")
	})

	dbPath := filepath.Join(tmpDir, "infra.db")

	db, err := database.NewDB(database.Option{Driver: "sqlite3", DSN: testDSN(dbPath)})
	Expect(err).NotTo(HaveOccurred(), "connect test db")
	Expect(migrate.Execute(context.Background(), db.DB, migrate.FS, "scripts")).To(Succeed(), "migrate test db")
	Expect(db.Close()).To(Succeed(), "close migration connection")

	return []byte(dbPath)
}, func(dbPath []byte) {
	db, err := database.NewDB(database.Option{
		Driver: "sqlite3",
		DSN:    testDSN(string(dbPath)),
	})
	Expect(err).NotTo(HaveOccurred(), "connect test db")
	testDB = db
	DeferCleanup(func() {
		Expect(db.Close()).To(Succeed())
	})
})

// entity 的内存缓存不参与事务回滚，每个 spec 都换一个新的缓存，
// 避免 spec 之间通过缓存读到已被回滚的数据
var _ = BeforeEach(func() {
	entity.DefaultCacher = cache.NewMemoryCache()
})

// beginTestTx 开启 spec 级事务，spec 结束后（含失败）自动回滚
//
// 事务不能绑定 SpecContext：Ginkgo 在每个 node 返回时就取消该节点的 ctx，
// database/sql 会据此自动回滚事务并废弃连接，sqlite :memory: 的库也会随连接
// 一起消失，后续 spec 会报 "no such table"。
func beginTestTx() *sqlx.Tx {
	GinkgoHelper()

	tx, err := testDB.Beginx()
	Expect(err).NotTo(HaveOccurred(), "begin test transaction")

	DeferCleanup(func() {
		Expect(tx.Rollback()).To(Succeed(), "rollback test transaction")
	})

	return tx
}
