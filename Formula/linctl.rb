class Linctl < Formula
  desc "Comprehensive command-line interface for Linear's API"
  homepage "https://github.com/dorkitude/linctl"
  url "https://github.com/dorkitude/linctl/archive/refs/tags/v0.1.5.tar.gz"
  sha256 "fddd9b17f07279663fa5f23e12e56b9f025064ba824a4c24c433ea8699affefb"
  license "MIT"
  head "https://github.com/dorkitude/linctl.git", branch: "master"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X github.com/dorkitude/linctl/cmd.version=#{version}")
  end

  test do
    # Test version output
    assert_match "linctl version #{version}", shell_output("#{bin}/linctl --version")
    
    # Test help command
    assert_match "A comprehensive CLI tool for Linear", shell_output("#{bin}/linctl --help")
  end
end
