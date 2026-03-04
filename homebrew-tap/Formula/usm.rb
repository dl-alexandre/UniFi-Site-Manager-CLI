class Usm < Formula
  desc "UniFi Site Manager CLI for cloud API management"
  homepage "https://github.com/dl-alexandre/UniFi-Site-Manager-CLI"
  version "0.0.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/download/v#{version}/usm_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_ARM64"
    else
      url "https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/download/v#{version}/usm_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/download/v#{version}/usm_linux_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
    else
      url "https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/download/v#{version}/usm_linux_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"
    end
  end

  def install
    bin.install "usm"
    
    # Install shell completions if available
    bash_completion.install "completions/usm.bash" => "usm" if File.exist?("completions/usm.bash")
    zsh_completion.install "completions/_usm" => "_usm" if File.exist?("completions/_usm")
    fish_completion.install "completions/usm.fish" if File.exist?("completions/usm.fish")
  end

  test do
    system "#{bin}/usm", "--version"
  end
end
