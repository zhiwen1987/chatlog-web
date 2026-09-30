# 模板目录

这些文件是交接/派工/部署/验收记录格式，不是产品程序配置。

- [HANDOFF](HANDOFF.md)：真实代码快照、三槽状态、外部副作用和继续入口。
- [WORK_ITEM](WORK_ITEM.yaml)：单工作项的owner、合同、路径和测试。
- [DEPLOYMENT_RECORD](DEPLOYMENT_RECORD.yaml)：桌面/客户服务端/发行方的真实部署事实。
- [RELEASE_MANIFEST](RELEASE_MANIFEST.example.json)：实际发行物与证据；空artifacts/images禁止当正式发行清单。
- [IMPLEMENTATION_STATUS](IMPLEMENTATION_STATUS.example.json)：所有模块从unknown开始。
- [PRODUCT_ACCEPTANCE](PRODUCT_ACCEPTANCE.example.json)：从开发规则提取的编号表格；不是取代正文的全量断言实现。未纳入表格的DOC-01～DOC-08及正文要求仍须验收。
- [MCP_TOOL_CONTRACT](MCP_TOOL_CONTRACT.yaml)：单业务工具的权限/合同/错误与验收设计。
- [KNOWN_ISSUES](KNOWN_ISSUES.md)：缺陷、阻塞与修复证据。

复制到真实工程后逐项填写，不在此样例上批量把unknown/not_tested改成pass。不要把示例被解析成功当作业务验收成功。
