class DevhelperCli < Formula
  desc "A comprehensive command-line interface for ShieldDev operations"
  homepage "https://github.com/lirtsman/devhelper-cli"
  url "https://github.com/lirtsman/devhelper-cli/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "971d6150e4b239d8261c3a6b99f14ca7789fd9c890ebbebd289ef2808a9ad9a1"
  license "Apache-2.0"
  head "https://github.com/lirtsman/devhelper-cli.git", branch: "main"

  depends_on "go" => :build

  def install
    # Get current time for build date
    build_date = Time.now.utc.strftime("%Y-%m-%dT%H:%M:%SZ")

    # Build from source
    system "go", "build", 
           "-ldflags", "-X github.com/lirtsman/devhelper-cli/cmd.Version=#{version} -X github.com/lirtsman/devhelper-cli/cmd.BuildDate=#{build_date} -X github.com/lirtsman/devhelper-cli/cmd.Commit=HEAD",
           "-o", "devhelper-cli"
    
    # Install the binary
    bin.install "devhelper-cli"
  end

  test do
    # Add a test to ensure the binary works correctly
    assert_match "ShieldDev CLI", shell_output("#{bin}/devhelper-cli --help", 0)
  end
end # End of file
