class BridgeAgent < Formula
  desc "BridgeAIChat remote device agent"
  homepage "https://bridgeai.chat"
  license "MIT"
  version "v0.1.0-beta.1"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/bridge-ai-chat/bridge-agent/releases/download/v0.1.0-beta.1/bridge-agent_v0.1.0-beta.1_darwin_arm64.tar.gz"
      sha256 "9ea1468875e97db3b198ba060013d4d94e3f3bcbb2299a7a44c2509838be3025"
    else
      url "https://github.com/bridge-ai-chat/bridge-agent/releases/download/v0.1.0-beta.1/bridge-agent_v0.1.0-beta.1_darwin_amd64.tar.gz"
      sha256 "dfdeb11c0579b058df49667eae92e17fb0b5665088ac7845922be0b63dc34df1"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/bridge-ai-chat/bridge-agent/releases/download/v0.1.0-beta.1/bridge-agent_v0.1.0-beta.1_linux_arm64.tar.gz"
      sha256 "f7e4d9acce824abe382dc63e4096edf6cf6924933e2d3cdf4a207062da2a5e8d"
    else
      url "https://github.com/bridge-ai-chat/bridge-agent/releases/download/v0.1.0-beta.1/bridge-agent_v0.1.0-beta.1_linux_amd64.tar.gz"
      sha256 "5f877368363d130354f87e75acbacfebaba13846fa8102b7589cfdbc6293a0a6"
    end
  end

  def install
    bin.install "bridge-agent"
    pkgshare.install "agent.yaml.example"
    pkgshare.install "install-agent.sh"
  end

  def caveats
    <<~EOS
      Run the installer to create ~/.bridge-agent/agent.yaml:
        bash "#{pkgshare}/install-agent.sh"

      bridge-agent requires tmux and a supported AI CLI on this device.
    EOS
  end

  test do
    assert_match "config", shell_output("#{bin}/bridge-agent -h 2>&1", 1)
  end
end
