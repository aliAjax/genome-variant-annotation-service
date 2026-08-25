# Genome Variant Annotation

纯Go实现的基因组变异规范化、参考数据版本管理和批量注释服务。服务流式解析VCF，拆分多等位记录，生成稳定变异键，并基于已发布的数据集执行区间和频率注释。默认内存模式无需外部依赖。

## 启动与验证

`HTTP_ADDR=:28088 go run ./cmd/server`

运行`go run ./cmd/simulator -base http://127.0.0.1:28088`可创建并发布参考数据集、提交注释任务、查询结果和执行规范化预览。服务提供`/healthz`、`/readyz`和`/metrics`。

工程目录包含`cmd`、`internal`、`api`、`configs`、`migrations`、`deploy`和`scripts`。核心领域按domain/application/adapter/infrastructure职责拆分，通过接口注入时钟、存储、参考序列和任务队列。

非测试Go源码不少于2400行，排除测试、生成代码、依赖、锁文件、SQL、配置、脚本和构建产物。建议使用`gofmt -w . && go test -race ./... && go vet ./... && go build ./...`验收。
