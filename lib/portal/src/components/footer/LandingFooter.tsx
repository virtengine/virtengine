/**
 * Landing Footer
 * VE Portal Landing Page
 */
import * as React from "react";

export interface FooterLinkGroup {
  title: string;
  links: Array<{ label: string; href: string }>;
}

export interface LandingFooterProps {
  className?: string;
  linkGroups?: FooterLinkGroup[];
  legalLinks?: Array<{ label: string; href: string }>;
  socialLinks?: Array<{ label: string; href: string }>;
}

const defaultLinkGroups: FooterLinkGroup[] = [
  {
    title: "Platform",
    links: [
      { label: "Marketplace", href: "/marketplace" },
      { label: "Providers", href: "/providers" },
      { label: "Pricing", href: "/pricing" },
      { label: "Docs", href: "/docs" },
    ],
  },
  {
    title: "Developers",
    links: [
      { label: "SDK", href: "/sdk" },
      { label: "API Status", href: "/status" },
      { label: "Changelog", href: "/changelog" },
      { label: "Governance", href: "/governance" },
    ],
  },
  {
    title: "Company",
    links: [
      { label: "About", href: "/about" },
      { label: "Careers", href: "/careers" },
      { label: "Press", href: "/press" },
      { label: "Contact", href: "/contact" },
    ],
  },
];

const defaultLegalLinks = [
  { label: "Privacy", href: "/privacy" },
  { label: "Terms", href: "/terms" },
  { label: "Security", href: "/security" },
];

const defaultSocialLinks = [
  { label: "X", href: "https://x.com/virtengine" },
  { label: "GitHub", href: "https://github.com/virtengine" },
  { label: "Discord", href: "https://discord.gg/virtengine" },
];

export function LandingFooter({
  className = "",
  linkGroups = defaultLinkGroups,
  legalLinks = defaultLegalLinks,
  socialLinks = defaultSocialLinks,
}: LandingFooterProps): JSX.Element {
  return (
    <footer className={`ve-footer ${className}`}>
      <div className="ve-footer__content">
        <div className="ve-footer__brand">
          <span className="ve-footer__logo">VE</span>
          <p>
            VirtEngine connects customers and providers through a transparent,
            verifiable compute marketplace.
          </p>
          <div className="ve-footer__socials">
            {socialLinks.map((link) => (
              <a key={link.label} href={link.href} aria-label={link.label}>
                {link.label}
              </a>
            ))}
          </div>
        </div>

        <div className="ve-footer__links">
          {linkGroups.map((group) => (
            <div key={group.title}>
              <span className="ve-footer__title">{group.title}</span>
              <ul>
                {group.links.map((link) => (
                  <li key={link.label}>
                    <a href={link.href}>{link.label}</a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </div>

      <div className="ve-footer__legal">
        <span>
          Copyright {new Date().getFullYear()} VirtEngine. All rights reserved.
        </span>
        <div className="ve-footer__legal-links">
          {legalLinks.map((link) => (
            <a key={link.label} href={link.href}>
              {link.label}
            </a>
          ))}
        </div>
      </div>

      <style>{footerStyles}</style>
    </footer>
  );
}

const footerStyles = `
  .ve-footer {
    background: #020617;
    color: #e2e8f0;
    padding: 72px 0 40px;
  }

  .ve-footer__content {
    width: min(1120px, 90vw);
    margin: 0 auto 32px;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 40px;
  }

  .ve-footer__brand {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .ve-footer__logo {
    width: 48px;
    height: 48px;
    border-radius: 14px;
    display: grid;
    place-items: center;
    background: linear-gradient(135deg, #38bdf8, #2563eb);
    font-weight: 700;
    color: #0f172a;
    font-family: "Space Grotesk", "Manrope", sans-serif;
  }

  .ve-footer__brand p {
    margin: 0;
    color: rgba(226, 232, 240, 0.75);
  }

  .ve-footer__socials {
    display: flex;
    gap: 12px;
  }

  .ve-footer__socials a {
    color: #38bdf8;
    text-decoration: none;
    font-weight: 600;
    font-size: 0.85rem;
  }

  .ve-footer__links {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
    gap: 24px;
  }

  .ve-footer__title {
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.18em;
    font-size: 0.75rem;
  }

  .ve-footer__links ul {
    list-style: none;
    padding: 0;
    margin: 12px 0 0;
    display: grid;
    gap: 8px;
  }

  .ve-footer__links a {
    color: rgba(226, 232, 240, 0.75);
    text-decoration: none;
  }

  .ve-footer__links a:hover {
    color: #f8fafc;
  }

  .ve-footer__legal {
    width: min(1120px, 90vw);
    margin: 0 auto;
    display: flex;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 16px;
    font-size: 0.85rem;
    color: rgba(226, 232, 240, 0.6);
  }

  .ve-footer__legal-links {
    display: flex;
    gap: 16px;
  }

  .ve-footer__legal-links a {
    color: rgba(226, 232, 240, 0.6);
    text-decoration: none;
  }

  .ve-footer__legal-links a:hover {
    color: #f8fafc;
  }
`;
