class DevhelperCli < Formula
  desc "A comprehensive command-line interface for ShieldDev operations"
  homepage "https://github.com/lirtsman/devhelper-cli"
  license "Apache-2.0"
  
  # Dynamic version based on latest release
  # This will automatically use the latest version
  version "v0.1.1"
  
  # Dynamic URLs for different architectures
  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/lirtsman/devhelper-cli/releases/latest/download/devhelper-cli-darwin-arm64"
      sha256 "PUT_ARM64_CHECKSUM_HERE" # You'll need to replace this with the actual SHA256 checksum
    else
      url "https://github.com/lirtsman/devhelper-cli/releases/latest/download/devhelper-cli-darwin-amd64"
      sha256 "PUT_AMD64_CHECKSUM_HERE" # You'll need to replace this with the actual SHA256 checksum
    end
  end
  
  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/lirtsman/devhelper-cli/releases/latest/download/devhelper-cli-linux-arm64"
      sha256 "PUT_ARM64_CHECKSUM_HERE" # You'll need to replace this with the actual SHA256 checksum
    else
      url "https://github.com/lirtsman/devhelper-cli/releases/latest/download/devhelper-cli-linux-amd64"
      sha256 "PUT_AMD64_CHECKSUM_HERE" # You'll need to replace this with the actual SHA256 checksum
    end
  end
  
  # Tell Homebrew that this is a binary, not a build-from-source formula
  bottle :unneeded

  def install
    # The downloaded file is already a binary, so we just need to rename it and make it executable
    bin.install Dir["devhelper-cli-*"].first => "devhelper-cli"
  end

  test do
    # Add a test to ensure the binary works correctly
    assert_match "ShieldDev CLI", shell_output("#{bin}/devhelper-cli --help")
  end
end 