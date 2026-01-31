# Directory Cleanup Plan

## Current Issues

### 1. Root Directory Clutter
临时文件和测试文件散落在根目录：
- hello.txt, test2.txt - 测试文件
- simple.go - 临时代码
- coverage.out - 覆盖率文件
- swarm.log - 日志文件
- swarm - 编译后的二进制

### 2. Scripts Not Organized
脚本文件散落：
- add_test_tasks.sh
- start_test_swarm.sh
- organize-repo.sh
- run-all-tests.sh

### 3. Reports in Multiple Locations
报告文件在多个位置：
- 根目录：TEST_RESULTS.md, UNIT_TEST_SUMMARY.md
- docs/reports/：大量报告
- docs/：部分文档

### 4. Missing Directories
缺少的目录结构：
- logs/ - 日志文件
- test/coverage/ - 测试覆盖率
- test/fixtures/ - 测试数据

## Proposed Structure

```
claude-swarm/
├── README.md
├── README_ZH.md
├── LICENSE
├── go.mod
├── go.sum
├── config.yaml.example
│
├── bin/                    # 编译后的二进制
│   └── swarm
│
├── cmd/                    # 命令行入口
│   └── swarm/
│
├── internal/               # 内部包
│   ├── config/
│   ├── models/
│   └── utils/
│
├── pkg/                    # 公共包
│   ├── analyzer/
│   ├── commander/
│   ├── controller/
│   └── ...
│
├── scripts/                # 所有脚本统一存放
│   ├── test/               # 测试相关脚本
│   │   ├── run-all-tests.sh
│   │   ├── add-test-tasks.sh
│   │   └── start-test-swarm.sh
│   ├── build/              # 构建脚本
│   └── utils/              # 工具脚本
│       └── organize-repo.sh
│
├── test/                   # 测试相关文件
│   ├── coverage/           # 覆盖率报告
│   │   └── coverage.out
│   ├── fixtures/           # 测试数据
│   └── integration/        # 集成测试
│
├── logs/                   # 日志文件
│   └── swarm.log
│
├── docs/                   # 文档
│   ├── README.md           # 文档索引
│   ├── DIRECTORY_STRUCTURE.md
│   │
│   ├── guides/             # 使用指南
│   │   ├── GETTING_STARTED.md
│   │   ├── USER_GUIDE.md
│   │   ├── CONFIG_GUIDE.md
│   │   └── ...
│   │
│   ├── architecture/       # 架构文档
│   │   └── full-plan.md
│   │
│   ├── reports/            # 所有报告统一存放
│   │   ├── test/           # 测试报告
│   │   │   ├── P0_TEST_REPORT.md
│   │   │   ├── UNIT_TEST_SUMMARY.md
│   │   │   └── TEST_RESULTS.md
│   │   ├── bugfix/         # Bug修复报告
│   │   │   ├── CONFIRMATION_ISSUES_REPORT.md
│   │   │   └── P0_COMPLETION_SUMMARY.md
│   │   └── improvements/   # 改进报告
│   │       └── CODE_QUALITY_IMPROVEMENTS.md
│   │
│   └── api/                # API文档 (如果需要)
│
├── .worktrees/             # Git worktrees (忽略)
└── .archive/               # 归档文件 (忽略)
```

## Actions to Take

### 1. Create Missing Directories
```bash
mkdir -p logs
mkdir -p test/coverage
mkdir -p test/fixtures
mkdir -p test/integration
mkdir -p scripts/test
mkdir -p scripts/build
mkdir -p scripts/utils
mkdir -p docs/reports/test
mkdir -p docs/reports/bugfix
mkdir -p docs/reports/improvements
```

### 2. Move Files

#### Scripts
```bash
mv add_test_tasks.sh scripts/test/
mv start_test_swarm.sh scripts/test/
mv run-all-tests.sh scripts/test/
mv organize-repo.sh scripts/utils/
```

#### Test & Coverage
```bash
mv coverage.out test/coverage/
```

#### Logs
```bash
mv swarm.log logs/
mv *.log logs/ 2>/dev/null || true
```

#### Binary
```bash
mv swarm bin/ 2>/dev/null || true
```

#### Reports
```bash
mv TEST_RESULTS.md docs/reports/test/
mv UNIT_TEST_SUMMARY.md docs/reports/test/
mv DIRECTORY_STRUCTURE.md docs/
```

#### Organize existing reports
```bash
cd docs/reports/
mv P0_*.md test/
mv *TEST*.md test/
mv CONFIRMATION_ISSUES_REPORT.md bugfix/
mv CODE_QUALITY_IMPROVEMENTS.md improvements/
```

### 3. Delete Temporary Files
```bash
rm -f hello.txt test2.txt simple.go
rm -f test-*.go  # 如果有临时测试文件
```

### 4. Update .gitignore
```gitignore
# Binaries
/bin/swarm
/swarm

# Test coverage
/test/coverage/
*.out
coverage.html

# Logs
/logs/
*.log

# Temporary files
*.tmp
*.temp
hello.txt
test*.txt
simple.go

# OS files
.DS_Store
Thumbs.db

# IDE
.vscode/
.idea/
*.swp
*.swo

# Build artifacts
dist/
build/
```

### 5. Update Documentation
- Update README.md with new structure
- Create docs/README.md as documentation index
- Update references to moved files

## Benefits

1. ✅ **清晰的目录结构** - 每个目录有明确的用途
2. ✅ **更好的可维护性** - 文件易于查找
3. ✅ **专业性** - 符合Go项目标准结构
4. ✅ **版本控制友好** - 临时文件不会被误提交
5. ✅ **测试组织** - 测试文件和覆盖率分离
6. ✅ **文档组织** - 报告分类清晰

## Next Steps

1. Review this plan
2. Execute cleanup script
3. Update all references
4. Commit changes
