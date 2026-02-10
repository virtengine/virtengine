'use client';

import Link from 'next/link';
import { useTranslation } from 'react-i18next';

export function Footer() {
  const { t } = useTranslation();

  return (
    <footer className="border-t border-border bg-background">
      <div className="container py-8">
        <div className="grid gap-8 md:grid-cols-4">
          <div>
            <div className="flex items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary font-bold text-primary-foreground">
                V
              </div>
              <span className="font-semibold">{t('VirtEngine')}</span>
            </div>
            <p className="mt-4 text-sm text-muted-foreground">
              {t(
                'Decentralized cloud computing marketplace with ML-powered identity verification.'
              )}
            </p>
          </div>

          <div>
            <h3 className="font-semibold">{t('Product')}</h3>
            <ul className="mt-4 space-y-2 text-sm">
              <li>
                <Link href="/marketplace" className="text-muted-foreground hover:text-foreground">
                  {t('Marketplace')}
                </Link>
              </li>
              <li>
                <Link href="/hpc/jobs" className="text-muted-foreground hover:text-foreground">
                  {t('HPC Computing')}
                </Link>
              </li>
              <li>
                <Link href="/identity" className="text-muted-foreground hover:text-foreground">
                  {t('Identity (VEID)')}
                </Link>
              </li>
              <li>
                <Link
                  href="/provider/dashboard"
                  className="text-muted-foreground hover:text-foreground"
                >
                  {t('Become a Provider')}
                </Link>
              </li>
            </ul>
          </div>

          <div>
            <h3 className="font-semibold">{t('Resources')}</h3>
            <ul className="mt-4 space-y-2 text-sm">
              <li>
                <a
                  href="https://docs.virtengine.io"
                  className="text-muted-foreground hover:text-foreground"
                >
                  {t('Documentation')}
                </a>
              </li>
              <li>
                <a
                  href="https://github.com/virtengine"
                  className="text-muted-foreground hover:text-foreground"
                >
                  {t('GitHub')}
                </a>
              </li>
              <li>
                <Link
                  href="/governance/proposals"
                  className="text-muted-foreground hover:text-foreground"
                >
                  {t('Governance')}
                </Link>
              </li>
              <li>
                <a
                  href="https://status.virtengine.com"
                  className="text-muted-foreground hover:text-foreground"
                >
                  {t('Status')}
                </a>
              </li>
            </ul>
          </div>

          <div>
            <h3 className="font-semibold">{t('Community')}</h3>
            <ul className="mt-4 space-y-2 text-sm">
              <li>
                <a
                  href="https://discord.gg/virtengine"
                  className="text-muted-foreground hover:text-foreground"
                >
                  {t('Discord')}
                </a>
              </li>
              <li>
                <a
                  href="https://twitter.com/virtengine"
                  className="text-muted-foreground hover:text-foreground"
                >
                  {t('Twitter')}
                </a>
              </li>
              <li>
                <a
                  href="https://forum.virtengine.com"
                  className="text-muted-foreground hover:text-foreground"
                >
                  {t('Forum')}
                </a>
              </li>
              <li>
                <a
                  href="https://blog.virtengine.com"
                  className="text-muted-foreground hover:text-foreground"
                >
                  {t('Blog')}
                </a>
              </li>
            </ul>
          </div>
        </div>

        <div className="mt-8 flex flex-col items-center justify-between gap-4 border-t border-border pt-8 md:flex-row">
          <p className="text-sm text-muted-foreground">
            {t('© {{year}} VirtEngine. All rights reserved.', {
              year: new Date().getFullYear(),
            })}
          </p>
          <div className="flex gap-6 text-sm text-muted-foreground">
            <a href="https://virtengine.com/privacy" className="hover:text-foreground">
              {t('Privacy Policy')}
            </a>
            <a href="https://virtengine.com/terms" className="hover:text-foreground">
              {t('Terms of Service')}
            </a>
            <a href="https://virtengine.com/cookies" className="hover:text-foreground">
              {t('Cookie Policy')}
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
}
