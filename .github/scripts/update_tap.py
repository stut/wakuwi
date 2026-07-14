#!/usr/bin/env python3
"""Update the wakuwi formula in the homebrew-tools tap."""
import base64
import subprocess
import sys

version, arm64_sha, amd64_sha, linux_arm64_sha, linux_amd64_sha = sys.argv[1:6]

formula = f"""\
class Wakuwi < Formula
  desc "Lightweight, read-only Kubernetes UI"
  homepage "https://github.com/stut/wakuwi"
  version "{version}"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/stut/wakuwi/releases/download/v#{{version}}/wakuwi-darwin-arm64"
      sha256 "{arm64_sha}"
    end
    on_intel do
      url "https://github.com/stut/wakuwi/releases/download/v#{{version}}/wakuwi-darwin-amd64"
      sha256 "{amd64_sha}"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/stut/wakuwi/releases/download/v#{{version}}/wakuwi-linux-arm64"
      sha256 "{linux_arm64_sha}"
    end
    on_intel do
      url "https://github.com/stut/wakuwi/releases/download/v#{{version}}/wakuwi-linux-amd64"
      sha256 "{linux_amd64_sha}"
    end
  end

  def install
    bin.install Dir["wakuwi-*"].first => "wakuwi"
  end

  service do
    run [opt_bin/"wakuwi"]
    keep_alive true
    log_path var/"log/wakuwi.log"
    error_log_path var/"log/wakuwi.log"
  end

  def caveats
    <<~EOS
      To start wakuwi as a background service:
        brew services start stut/tools/wakuwi

      Or run it manually:
        wakuwi

      Then open http://localhost:9753 in your browser.
    EOS
  end

  test do
    assert_match version.to_s, shell_output("#{{bin}}/wakuwi --version 2>&1", 1)
  end
end
"""

file_sha = subprocess.check_output(
    ["gh", "api", "repos/stut/homebrew-tools/contents/Formula/wakuwi.rb", "--jq", ".sha"],
    text=True,
).strip()

content = base64.b64encode(formula.encode()).decode()

subprocess.run(
    [
        "gh", "api", "--method", "PUT",
        "repos/stut/homebrew-tools/contents/Formula/wakuwi.rb",
        "--field", f"message=chore: update wakuwi to v{version}",
        "--field", f"sha={file_sha}",
        "--field", f"content={content}",
    ],
    check=True,
)

print(f"Updated homebrew-tools formula to v{version}")
