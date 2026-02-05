# Homebrew formula for SpecLedger CLI
# Place this file in a homebrew tap repository, e.g.,:
#   github.com/specledger/homebrew-specledger

class Specledger < Formula
  desc "Unified CLI for project bootstrap and specification dependency management"
  homepage "https://github.com/specledger/specledger"
  url "https://github.com/specledger/specledger/releases/download/v1.0.0/specledger_1.0.0_darwin_amd64.tar.gz"
  sha256 "abc123def456..."

  bottle do
    sha256 cellar: :any_skip_relocation, arm64_sonoma:   "a1b2c3d4e5f6..."
    sha256 cellar: :any_skip_relocation, arm64_ventura:  "a1b2c3d4e5f6..."
    sha256 cellar: :any_skip_relocation, arm64_monterey: "a1b2c3d4e5f6..."
    sha256 cellar: :any_skip_relocation, sonoma:         "a1b2c3d4e5f6..."
    sha256 cellar: :any_skip_relocation, ventura:        "a1b2c3d4e5f6..."
    sha256 cellar: :any_skip_relocation, monterey:       "a1b2c3d4e5f6..."
    sha256 cellar: :any_skip_relocation, x86_64_linux:   "a1b2c3d4e5f6..."
  end

  depends_on "go" => :build

  def install
    # Extract the binary from the archive
    system "tar", "xzf", "#{tarball}", "-C", "#{buildpath}"

    # Install the binary
    bin.install "sl"
  end

  test do
    system bin/"sl", "version"
    assert_match(/specledger/, shell_output("#{bin}/sl --help"))
  end
end
