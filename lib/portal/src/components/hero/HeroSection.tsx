/**
 * Hero Section
 * VE Portal Landing Page
 */
import * as React from "react";

export interface HeroCTA {
  label: string;
  href?: string;
  onClick?: () => void;
  variant?: "primary" | "secondary";
}

export interface HeroSectionProps {
  headline?: string;
  subheadline?: string;
  primaryCta?: HeroCTA;
  secondaryCta?: HeroCTA;
  highlights?: string[];
  className?: string;
}

export function HeroSection({
  headline = "Decentralized Cloud Computing",
  subheadline = "Launch workloads on a permissionless network of verified providers. Elastic capacity, transparent pricing, and cryptographic trust baked in.",
  primaryCta = {
    label: "Browse Marketplace",
    href: "/marketplace",
    variant: "primary",
  },
  secondaryCta = {
    label: "Become Provider",
    href: "/providers/apply",
    variant: "secondary",
  },
  highlights = [
    "Multi-region compute pools",
    "On-chain settlement & auditability",
    "Confidential workloads ready",
  ],
  className = "",
}: HeroSectionProps): JSX.Element {
  return (
    <section className={`ve-hero ${className}`} aria-labelledby="ve-hero-title">
      <div className="ve-hero__background" aria-hidden="true">
        <div className="ve-hero__glow" />
        <div className="ve-hero__grid" />
        <div className="ve-hero__orb ve-hero__orb--one" />
        <div className="ve-hero__orb ve-hero__orb--two" />
        <div className="ve-hero__orb ve-hero__orb--three" />
      </div>

      <div className="ve-hero__content">
        <div className="ve-hero__text">
          <span className="ve-hero__eyebrow">VirtEngine Portal</span>
          <h1 id="ve-hero-title" className="ve-hero__headline">
            {headline}
          </h1>
          <p className="ve-hero__subheadline">{subheadline}</p>
          <div className="ve-hero__cta">
            <HeroButton cta={primaryCta} variant="primary" />
            <HeroButton cta={secondaryCta} variant="secondary" />
          </div>
          <ul className="ve-hero__highlights" aria-label="Key highlights">
            {highlights.map((item) => (
              <li key={item} className="ve-hero__highlight">
                <span className="ve-hero__highlight-dot" aria-hidden="true" />
                <span>{item}</span>
              </li>
            ))}
          </ul>
        </div>

        <div className="ve-hero__visual" aria-hidden="true">
          <div className="ve-hero__panel">
            <div className="ve-hero__panel-header">
              <span>Network pulse</span>
              <span className="ve-hero__panel-status">Live</span>
            </div>
            <div className="ve-hero__panel-metrics">
              <div>
                <span className="ve-hero__panel-label">Latency</span>
                <strong>42ms</strong>
              </div>
              <div>
                <span className="ve-hero__panel-label">Uptime</span>
                <strong>99.98%</strong>
              </div>
              <div>
                <span className="ve-hero__panel-label">Regions</span>
                <strong>23</strong>
              </div>
            </div>
            <div className="ve-hero__pulse">
              <span />
              <span />
              <span />
            </div>
          </div>
          <div className="ve-hero__nodes">
            <div className="ve-hero__node" />
            <div className="ve-hero__node" />
            <div className="ve-hero__node" />
            <div className="ve-hero__node" />
          </div>
        </div>
      </div>

      <style>{heroStyles}</style>
    </section>
  );
}

interface HeroButtonProps {
  cta: HeroCTA;
  variant: "primary" | "secondary";
}

function HeroButton({ cta, variant }: HeroButtonProps): JSX.Element {
  const className = `ve-hero__button ve-hero__button--${cta.variant ?? variant}`;
  if (cta.onClick) {
    return (
      <button type="button" onClick={cta.onClick} className={className}>
        {cta.label}
      </button>
    );
  }

  return (
    <a href={cta.href} className={className}>
      {cta.label}
    </a>
  );
}

const heroStyles = `
  .ve-hero {
    position: relative;
    overflow: hidden;
    padding: 96px 0 72px;
    color: #f8fafc;
    background: #05070d;
  }

  .ve-hero__background {
    position: absolute;
    inset: 0;
    pointer-events: none;
    overflow: hidden;
  }

  .ve-hero__glow {
    position: absolute;
    width: 720px;
    height: 720px;
    border-radius: 50%;
    top: -320px;
    left: -240px;
    background: radial-gradient(circle, rgba(56, 189, 248, 0.35), transparent 70%);
    filter: blur(0px);
  }

  .ve-hero__grid {
    position: absolute;
    inset: 0;
    background-image: linear-gradient(rgba(148, 163, 184, 0.08) 1px, transparent 1px),
      linear-gradient(90deg, rgba(148, 163, 184, 0.08) 1px, transparent 1px);
    background-size: 48px 48px;
    opacity: 0.4;
    mask-image: radial-gradient(circle at 30% 20%, black 0%, transparent 65%);
  }

  .ve-hero__orb {
    position: absolute;
    border-radius: 999px;
    opacity: 0.7;
    filter: blur(0px);
    animation: orbFloat 14s ease-in-out infinite;
  }

  .ve-hero__orb--one {
    width: 240px;
    height: 240px;
    right: 15%;
    top: 10%;
    background: radial-gradient(circle, rgba(14, 116, 144, 0.65), transparent 65%);
    animation-delay: -2s;
  }

  .ve-hero__orb--two {
    width: 320px;
    height: 320px;
    right: -120px;
    bottom: -160px;
    background: radial-gradient(circle, rgba(59, 130, 246, 0.5), transparent 70%);
    animation-delay: -6s;
  }

  .ve-hero__orb--three {
    width: 180px;
    height: 180px;
    left: 15%;
    bottom: -40px;
    background: radial-gradient(circle, rgba(16, 185, 129, 0.45), transparent 70%);
    animation-delay: -9s;
  }

  .ve-hero__content {
    position: relative;
    z-index: 2;
    width: min(1120px, 90vw);
    margin: 0 auto;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 48px;
    align-items: center;
  }

  .ve-hero__text {
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  .ve-hero__eyebrow {
    font-size: 0.85rem;
    text-transform: uppercase;
    letter-spacing: 0.32em;
    font-weight: 600;
    color: rgba(226, 232, 240, 0.7);
  }

  .ve-hero__headline {
    font-family: "Fraunces", "Playfair Display", "Georgia", serif;
    font-size: clamp(2.6rem, 4vw, 3.8rem);
    line-height: 1.05;
    margin: 0;
    letter-spacing: -0.02em;
  }

  .ve-hero__subheadline {
    font-family: "Space Grotesk", "Manrope", "Segoe UI", sans-serif;
    font-size: 1.1rem;
    color: rgba(226, 232, 240, 0.78);
    margin: 0;
    max-width: 520px;
  }

  .ve-hero__cta {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
    margin-top: 8px;
  }

  .ve-hero__button {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 12px 22px;
    border-radius: 999px;
    font-size: 0.95rem;
    font-weight: 600;
    text-decoration: none;
    transition: transform 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
    border: 1px solid transparent;
  }

  .ve-hero__button--primary {
    background: linear-gradient(135deg, #38bdf8, #2563eb);
    color: #020617;
    box-shadow: 0 12px 40px rgba(56, 189, 248, 0.35);
  }

  .ve-hero__button--secondary {
    background: rgba(15, 23, 42, 0.2);
    color: #e2e8f0;
    border-color: rgba(148, 163, 184, 0.35);
  }

  .ve-hero__button:hover {
    transform: translateY(-2px);
  }

  .ve-hero__highlights {
    list-style: none;
    padding: 0;
    margin: 12px 0 0;
    display: grid;
    gap: 10px;
    color: rgba(226, 232, 240, 0.82);
    font-size: 0.95rem;
  }

  .ve-hero__highlight {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .ve-hero__highlight-dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: linear-gradient(135deg, #22d3ee, #0ea5e9);
    box-shadow: 0 0 12px rgba(34, 211, 238, 0.7);
  }

  .ve-hero__visual {
    position: relative;
    display: flex;
    justify-content: center;
  }

  .ve-hero__panel {
    background: rgba(15, 23, 42, 0.7);
    border: 1px solid rgba(148, 163, 184, 0.2);
    border-radius: 24px;
    padding: 24px;
    width: min(360px, 90%);
    box-shadow: 0 20px 60px rgba(15, 23, 42, 0.6);
    backdrop-filter: blur(18px);
  }

  .ve-hero__panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.85rem;
    color: rgba(226, 232, 240, 0.7);
  }

  .ve-hero__panel-status {
    background: rgba(34, 197, 94, 0.18);
    color: #22c55e;
    padding: 4px 10px;
    border-radius: 999px;
    font-weight: 600;
    font-size: 0.75rem;
  }

  .ve-hero__panel-metrics {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 16px;
    margin-top: 24px;
    font-size: 0.85rem;
  }

  .ve-hero__panel-metrics strong {
    display: block;
    font-size: 1.15rem;
    color: #f8fafc;
  }

  .ve-hero__panel-label {
    color: rgba(226, 232, 240, 0.5);
  }

  .ve-hero__pulse {
    margin-top: 20px;
    display: flex;
    gap: 8px;
    justify-content: space-between;
  }

  .ve-hero__pulse span {
    display: block;
    height: 6px;
    flex: 1;
    background: linear-gradient(90deg, rgba(59, 130, 246, 0.2), rgba(34, 211, 238, 0.8));
    border-radius: 999px;
    animation: pulseBar 2.8s ease-in-out infinite;
  }

  .ve-hero__pulse span:nth-child(2) {
    animation-delay: 0.4s;
  }

  .ve-hero__pulse span:nth-child(3) {
    animation-delay: 0.8s;
  }

  .ve-hero__nodes {
    position: absolute;
    right: 10%;
    bottom: -16px;
    display: grid;
    grid-template-columns: repeat(2, 32px);
    gap: 16px;
  }

  .ve-hero__node {
    width: 32px;
    height: 32px;
    border-radius: 10px;
    background: rgba(15, 23, 42, 0.7);
    border: 1px solid rgba(59, 130, 246, 0.35);
    box-shadow: 0 6px 16px rgba(15, 23, 42, 0.4);
    animation: nodeGlow 3.8s ease-in-out infinite;
  }

  .ve-hero__node:nth-child(2) {
    animation-delay: 0.6s;
  }

  .ve-hero__node:nth-child(3) {
    animation-delay: 1.2s;
  }

  .ve-hero__node:nth-child(4) {
    animation-delay: 1.8s;
  }

  @keyframes orbFloat {
    0%, 100% { transform: translateY(0px); }
    50% { transform: translateY(18px); }
  }

  @keyframes pulseBar {
    0%, 100% { opacity: 0.4; transform: scaleX(0.6); }
    50% { opacity: 1; transform: scaleX(1); }
  }

  @keyframes nodeGlow {
    0%, 100% { box-shadow: 0 6px 16px rgba(15, 23, 42, 0.4); }
    50% { box-shadow: 0 12px 28px rgba(59, 130, 246, 0.5); }
  }

  @media (max-width: 860px) {
    .ve-hero {
      padding: 80px 0 64px;
    }

    .ve-hero__panel-metrics {
      grid-template-columns: repeat(2, 1fr);
    }

    .ve-hero__nodes {
      display: none;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .ve-hero__orb,
    .ve-hero__pulse span,
    .ve-hero__node {
      animation: none;
    }
  }
`;
