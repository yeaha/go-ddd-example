package infra

import (
	"context"
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
// 事务会一直占用唯一一条数据库连接，一旦有 spec 拿不到连接，没有超时就会
// 永久阻塞，因此这里显式设置上限，让失败快速暴露。
const specTimeout = SpecTimeout(10 * time.Second)

var testDB *sqlx.DB

func TestInfra(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Infra Suite")
}

var _ = BeforeSuite(func() {
	db, err := database.NewDB(database.Option{
		Driver: "sqlite3",
		DSN:    ":memory:",

		// sqlite :memory: 每个连接都是一个独立的库，限制成单连接才能保证
		// migrate 和被测代码操作的是同一个库。
		// 代价：spec 事务会一直占用这条唯一连接，所以 spec 内只能通过注入的
		// tx 访问数据库；直接使用 testDB 或再开嵌套事务都会阻塞。
		MaxOpenConns: 1,
	})
	Expect(err).NotTo(HaveOccurred(), "connect test db")
	testDB = db
	DeferCleanup(func() {
		Expect(db.Close()).To(Succeed())
	})

	Expect(migrate.Execute(context.Background(), db.DB, migrate.FS, "scripts")).To(Succeed(), "migrate test db")
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
