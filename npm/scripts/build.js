const { execSync } = require("child_process");
const fs = require("fs");
const path = require("path");

const ROOT = path.resolve(__dirname, "..");
const GO_ROOT = path.resolve(ROOT, ".."); // Go 项目根目录
const PLATFORM_DIR = path.join(ROOT, "platform");

const targets = [
  { goos: "windows", goarch: "amd64", dir: "51pm-win32-x64", bin: "51pm.exe" },
  { goos: "linux", goarch: "amd64", dir: "51pm-linux-x64", bin: "51pm" },
  { goos: "darwin", goarch: "amd64", dir: "51pm-darwin-x64", bin: "51pm" },
  { goos: "darwin", goarch: "arm64", dir: "51pm-darwin-arm64", bin: "51pm" },
];

// 清理
if (fs.existsSync(PLATFORM_DIR)) {
  fs.rmSync(PLATFORM_DIR, { recursive: true });
}

for (const target of targets) {
  const outDir = path.join(PLATFORM_DIR, target.dir);
  fs.mkdirSync(outDir, { recursive: true });

  const outPath = path.join(outDir, target.bin);
  console.log(`构建 ${target.goos}/${target.goarch} -> ${outPath}`);

  execSync(`go build -ldflags="-s -w" -o "${outPath}" .`, {
    cwd: GO_ROOT,
    env: {
      ...process.env,
      GOOS: target.goos,
      GOARCH: target.goarch,
      CGO_ENABLED: "0",
    },
    stdio: "inherit",
  });
}

console.log("\n所有平台构建完成！");
console.log("产物目录:", PLATFORM_DIR);
