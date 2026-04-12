class BridgeApp < Formula
  desc "BridgeAIChat gateway and hosted web UI"
  homepage "https://bridgeai.chat"
  license "MIT"
  version "v0.1.0-beta.1"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/bridge-ai-chat/bridge-app/releases/download/v0.1.0-beta.1/bridge-app_v0.1.0-beta.1_darwin_arm64.tar.gz"
      sha256 "b8a23601cfbea5cea19eb31ae6d109f1101a0dfa0eee4b928fa42c273f01c08b"
    else
      url "https://github.com/bridge-ai-chat/bridge-app/releases/download/v0.1.0-beta.1/bridge-app_v0.1.0-beta.1_darwin_amd64.tar.gz"
      sha256 "2280068b58df4553f26d2cb0360958ee175837a1abeabc10fdb295b9d13726d3"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/bridge-ai-chat/bridge-app/releases/download/v0.1.0-beta.1/bridge-app_v0.1.0-beta.1_linux_arm64.tar.gz"
      sha256 "d189747c16dbddc106753d2568d1fb9e339e637c2b7afee2274a753d7572cad0"
    else
      url "https://github.com/bridge-ai-chat/bridge-app/releases/download/v0.1.0-beta.1/bridge-app_v0.1.0-beta.1_linux_amd64.tar.gz"
      sha256 "1fbfa8b0d8e0242e6330c2a1e3f330272954fd5400f5adbb4d82b806183e2e3b"
    end
  end

  def install
    libexec.install "bridge-gateway"
    libexec.install "ui"
    libexec.install "install-app.sh"
    (bin/"bridge-gateway").write <<~EOS
      #!/usr/bin/env bash
      exec "#{libexec}/bridge-gateway" -ui-dist "#{libexec}/ui"
    EOS
  end

  def caveats
    <<~EOS
      Start the BridgeAIChat gateway with:
        bridge-gateway

      Then open:
        http://localhost:8080
    EOS
  end

  test do
    assert_match "gateway", shell_output("#{libexec}/bridge-gateway -h 2>&1", 1)
  end
end
