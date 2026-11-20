class Envy < Formula
  desc "End-to-end encrypted .env sync for developer teams"
  homepage "https://github.com/RovierrHQ/envy"
  version "0.2.2"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/RovierrHQ/envy/releases/download/v0.2.2/envy_0.2.2_darwin_arm64.tar.gz"
      sha256 "b92ce48c7a8f10938b316cb9e6314c6b9f489408e82433606aa156ab4a113c09"
    end
    on_intel do
      url "https://github.com/RovierrHQ/envy/releases/download/v0.2.2/envy_0.2.2_darwin_amd64.tar.gz"
      sha256 "701495e23e2f287678ef493abbf1f07890bc043b81d153859beefb6365d346da"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/RovierrHQ/envy/releases/download/v0.2.2/envy_0.2.2_linux_arm64.tar.gz"
      sha256 "13144c35447b039a4b0846922aae13fcb02dffe91f95204e0a2d5e145d2bf990"
    end
    on_intel do
      url "https://github.com/RovierrHQ/envy/releases/download/v0.2.2/envy_0.2.2_linux_amd64.tar.gz"
      sha256 "cc96af7e8e81f01e9729e513dd29796419b5c7277d42d30a111a3952aa33d54f"
    end
  end

  def install
    bin.install "envy"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/envy version")
  end
end
