class Envy < Formula
  desc "End-to-end encrypted .env sync for developer teams"
  homepage "https://github.com/RovierrHQ/envy"
  version "0.2.1"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/RovierrHQ/envy/releases/download/v0.2.1/envy_0.2.1_darwin_arm64.tar.gz"
      sha256 "aad17f6eca6facaa347070f2cf36d92d6aaa0cef246aeb87156eb120c403c1d6"
    end
    on_intel do
      url "https://github.com/RovierrHQ/envy/releases/download/v0.2.1/envy_0.2.1_darwin_amd64.tar.gz"
      sha256 "69b5ccb917d833570c2bfbff91ae26001bb4cb02fb1e1832556ac6c8c2be41a8"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/RovierrHQ/envy/releases/download/v0.2.1/envy_0.2.1_linux_arm64.tar.gz"
      sha256 "4cc9dfecdb2b6214a6b2a9800d32b9d86687f15c71d55fe3805f8390458ced84"
    end
    on_intel do
      url "https://github.com/RovierrHQ/envy/releases/download/v0.2.1/envy_0.2.1_linux_amd64.tar.gz"
      sha256 "ee17ccef17ec03327daf4baef4d38142c5d9285763955706c12fb23b34c709df"
    end
  end

  def install
    bin.install "envy"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/envy version")
  end
end
