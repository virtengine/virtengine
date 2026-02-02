import Link from 'next/link';

export function Footer() {
  return (
    <footer className="border-t border-border bg-background">
      <div className="container py-8">
        <div className="grid gap-8 md:grid-cols-4">
          <div>
            <div className="flex items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground font-bold">
                V
              </div>
              <span className="font-semibold">VirtEngine</span>
            </div>
            <p className="mt-4 text-sm text-muted-foreground">
              Decentralized cloud computing marketplace with ML-powered identity verification.
            </p>
          </div>

          <div>
            <h3 className="font-semibold">Product</h3>
            <ul className="mt-4 space-y-2 text-sm">
              <li>
                <Link href="/marketplace" className="text-muted-foreground hover:text-foreground">
                  Marketplace
                </Link>
              </li>
              <li>
                <Link href="/hpc/jobs" className="text-muted-foreground hover:text-foreground">
                  HPC Computing
                </Link>
              </li>
              <li>
                <Link href="/identity" className="text-muted-foreground hover:text-foreground">
                  Identity (VEID)
                </Link>
              </li>
              <li>
                <Link href="/provider/dashboard" className="text-muted-foreground hover:text-foreground">
                  Become a Provider
                </Link>
              </li>
            </ul>
          </div>

          <div>
            <h3 className="font-semibold">Resources</h3>
            <ul className="mt-4 space-y-2 text-sm">
              <li>
                <a href="https://docs.virtengine.io" className="text-muted-foreground hover:text-foreground">
                  Documentation
                </a>
              </li>
              <li>
                <a href="https://github.com/virtengine" className="text-muted-foreground hover:text-foreground">
                  GitHub
                </a>
              </li>
              <li>
                <Link href="/governance/proposals" className="text-muted-foreground hover:text-foreground">
                  Governance
                </Link>
              </li>
              <li>
                <a href="https://status.virtengine.com" className="text-muted-foreground hover:text-foreground">
                  Status
                </a>
              </li>
            </ul>
          </div>

          <div>
            <h3 className="font-semibold">Community</h3>
            <ul className="mt-4 space-y-2 text-sm">
              <li>
                <a href="https://discord.gg/virtengine" className="text-muted-foreground hover:text-foreground">
                  Discord
                </a>
              </li>
              <li>
                <a href="https://twitter.com/virtengine" className="text-muted-foreground hover:text-foreground">
                  Twitter
                </a>
              </li>
              <li>
                <a href="https://forum.virtengine.com" className="text-muted-foreground hover:text-foreground">
                  Forum
                </a>
              </li>
              <li>
                <a href="https://blog.virtengine.com" className="text-muted-foreground hover:text-foreground">
                  Blog
                </a>
              </li>
            </ul>
          </div>
        </div>

        <div className="mt-8 flex flex-col items-center justify-between gap-4 border-t border-border pt-8 md:flex-row">
          <p className="text-sm text-muted-foreground">
            © {new Date().getFullYear()} VirtEngine. All rights reserved.
          </p>
          <div className="flex gap-6 text-sm text-muted-foreground">
            <a href="https://virtengine.com/privacy" className="hover:text-foreground">
              Privacy Policy
            </a>
            <a href="https://virtengine.com/terms" className="hover:text-foreground">
              Terms of Service
            </a>
            <a href="https://virtengine.com/cookies" className="hover:text-foreground">
              Cookie Policy
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
}
