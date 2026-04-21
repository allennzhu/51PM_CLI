#!/usr/bin/env node

const { execFileSync } = require("child_process");
const path = require("path");
const os = require("os");
const fs = require("fs");

function getBinaryPath() {
  const platform = os.platform();
  const arch = os.arch();

  const platformMap = {
    "win32-x64": "51pm-win32-x64/51pm.exe",
    "linux-x64": "51pm-linux-x64/51pm",
    "darwin-x64": "51pm-darwin-x64/51pm",
    "darwin-arm64": "51pm-darwin-arm64/51pm",
  };

  const key = `${platform}-${arch}`;
  const relativePath = platformMap[key];

  if (!relativePath) {
    console.error(`不支持的平台: ${platform}-${arch}`);
    process.exit(1);
  }

  // 优先从 optionalDependencies 安装的平台包中查找
  const pkgName = `@anthropic-test/51pm-cli-${platform}-${arch}`;
  try {
    const pkgBin = require.resolve(path.join(pkgName, relativePath.split("/")[1]));
    if (fs.existsSync(pkgBin)) {
      return pkgBin;
    }
  } catch (e) {
    // 平台包未安装，回退到 platform/ 目录
  }

  // 回退：从本包 platform/ 目录查找
  const localBin = path.join(__dirname, "..", "platform", relativePath);
  if (fs.existsSync(localBin)) {
    return localBin;
  }

  console.error(
    `找不到 51pm 二进制文件。\n` +
    `请确认平台包 ${pkgName} 已正确安装，或 platform/ 目录包含对应二进制文件。`
  );
  process.exit(1);
}

const binPath = getBinaryPath();
const args = process.argv.slice(2);

try {
  execFileSync(binPath, args, { stdio: "inherit" });
} catch (e) {
  process.exit(e.status || 1);
}
