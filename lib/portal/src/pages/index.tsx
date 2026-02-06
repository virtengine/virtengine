/**
 * Portal Landing Page
 * VE Portal Landing Page
 */
import * as React from "react";
import { HeroSection } from "../components/hero";
import { StatsSection } from "../components/stats";
import { FeaturedOfferings } from "../components/offerings";
import { LandingFooter } from "../components/footer";
import { ProviderBadge } from "./marketplace/components/ProviderInfo";
import type { Offering } from "../../types/marketplace";

export interface LandingPageProps {
  className?: string;
  onOrderOffering?: (offering: Offering) => void;
}

export function LandingPage({
  className = "",
  onOrderOffering,
}: LandingPageProps): JSX.Element {
  return (
    <div className={`ve-landing ${className}`}>
      <HeroSection />
      <StatsSection />
      <FeaturedOfferings onOrder={onOrderOffering} />
      <HowItWorksSection />
      <ProviderSpotlightSection />
      <LandingFooter />
      <style>{landingStyles}</style>
    </div>
  );
}

function HowItWorksSection(): JSX.Element {
  const steps = [
    {
      title: "Discover verified providers",
      description:
        "Browse on-chain offerings with benchmarked performance, pricing, and compliance badges.",
    },
    {
      title: "Configure & reserve",
      description:
        "Select resources, lock pricing, and reserve capacity with escrow-backed settlement.",
    },
    {
      title: "Deploy securely",
      description:
        "Launch workloads with confidential computing and usage attestation baked in.",
    },
    {
      title: "Settle automatically",
      description:
        "Pay only for verified usage with transparent, traceable payouts to providers.",
    },
  ];

  const providerSteps = [
    "Register infrastructure and attest hardware",
    "Publish offerings with dynamic pricing",
    "Receive orders and provision instantly",
    "Earn rewards and reputation on-chain",
  ];

  return (
    <section className="ve-how" aria-labelledby="ve-how-title">
      <div className="ve-how__header">
        <h2 id="ve-how-title">How it works</h2>
        <p>
          From discovery to deployment, the marketplace keeps both customers and
          providers aligned.
        </p>
      </div>

      <div className="ve-how__grid">
        {steps.map((step, index) => (
          <div className="ve-how__card" key={step.title}>
            <span className="ve-how__step">{index + 1}</span>
            <h3>{step.title}</h3>
            <p>{step.description}</p>
          </div>
        ))}
      </div>

      <div className="ve-how__journeys">
        <div className="ve-how__journey">
          <h3>Customer journey</h3>
          <ol>
            {steps.map((step, index) => (
              <li key={step.title}>
                <span>{index + 1}</span>
                {step.title}
              </li>
            ))}
          </ol>
        </div>
        <div className="ve-how__journey">
          <h3>Provider journey</h3>
          <ol>
            {providerSteps.map((step, index) => (
              <li key={step}>
                <span>{index + 1}</span>
                {step}
              </li>
            ))}
          </ol>
        </div>
      </div>
    </section>
  );
}

function ProviderSpotlightSection(): JSX.Element {
  const providers = [
    {
      name: "NovaGrid",
      address: "ve1nova",
      reliability: 98,
      capacity: "18k CPU cores",
      highlights: [
        "SOC2 + ISO27001",
        "Tier-1 data centers",
        "Confidential GPU pools",
      ],
    },
    {
      name: "Fluxline Compute",
      address: "ve1flux",
      reliability: 94,
      capacity: "12k CPU cores",
      highlights: ["99.99% uptime", "Multi-region failover", "Green energy"],
    },
    {
      name: "Helios Array",
      address: "ve1helios",
      reliability: 91,
      capacity: "9k CPU cores",
      highlights: ["Low latency mesh", "HPC-ready clusters", "Burst pricing"],
    },
  ];

  return (
    <section className="ve-spotlight" aria-labelledby="ve-spotlight-title">
      <div className="ve-spotlight__header">
        <h2 id="ve-spotlight-title">Provider spotlight</h2>
        <p>High-trust providers delivering capacity at enterprise scale.</p>
      </div>

      <div className="ve-spotlight__grid">
        {providers.map((provider) => (
          <div className="ve-spotlight__card" key={provider.name}>
            <ProviderBadge
              providerName={provider.name}
              providerAddress={provider.address}
              reliabilityScore={provider.reliability}
              isVerified
            />
            <div className="ve-spotlight__capacity">
              <span>Capacity</span>
              <strong>{provider.capacity}</strong>
            </div>
            <ul>
              {provider.highlights.map((item) => (
                <li key={item}>{item}</li>
              ))}
            </ul>
            <a
              className="ve-spotlight__cta"
              href={`/providers/${provider.address}`}
            >
              View provider profile
            </a>
          </div>
        ))}
      </div>
    </section>
  );
}

const landingStyles = `
  @import url("https://fonts.googleapis.com/css2?family=Fraunces:wght@600;700&family=Space+Grotesk:wght@400;500;600;700&display=swap");

  .ve-landing {
    font-family: "Space Grotesk", "Manrope", "Segoe UI", sans-serif;
    background: #0f172a;
  }

  .ve-how {
    padding: 96px 0;
    background: linear-gradient(180deg, #f8fafc 0%, #e2e8f0 100%);
    color: #0f172a;
  }

  .ve-how__header {
    width: min(1120px, 90vw);
    margin: 0 auto 40px;
  }

  .ve-how__header h2 {
    font-size: clamp(1.8rem, 2.4vw, 2.4rem);
    margin: 0 0 8px;
  }

  .ve-how__header p {
    margin: 0;
    color: #475569;
  }

  .ve-how__grid {
    width: min(1120px, 90vw);
    margin: 0 auto;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 20px;
  }

  .ve-how__card {
    background: #ffffff;
    border-radius: 20px;
    padding: 24px;
    box-shadow: 0 20px 40px rgba(15, 23, 42, 0.1);
  }

  .ve-how__card h3 {
    margin: 12px 0 8px;
  }

  .ve-how__card p {
    margin: 0;
    color: #475569;
  }

  .ve-how__step {
    width: 36px;
    height: 36px;
    border-radius: 12px;
    display: grid;
    place-items: center;
    background: #0f172a;
    color: #f8fafc;
    font-weight: 700;
  }

  .ve-how__journeys {
    width: min(1120px, 90vw);
    margin: 48px auto 0;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
    gap: 24px;
  }

  .ve-how__journey {
    background: #0f172a;
    color: #e2e8f0;
    border-radius: 18px;
    padding: 24px;
  }

  .ve-how__journey h3 {
    margin: 0 0 16px;
  }

  .ve-how__journey ol {
    margin: 0;
    padding-left: 0;
    list-style: none;
    display: grid;
    gap: 12px;
  }

  .ve-how__journey li {
    display: flex;
    gap: 12px;
    align-items: center;
    font-size: 0.95rem;
  }

  .ve-how__journey li span {
    width: 28px;
    height: 28px;
    border-radius: 8px;
    display: grid;
    place-items: center;
    background: rgba(56, 189, 248, 0.2);
    color: #38bdf8;
    font-weight: 700;
  }

  .ve-spotlight {
    padding: 96px 0;
    background: #0f172a;
    color: #f8fafc;
  }

  .ve-spotlight__header {
    width: min(1120px, 90vw);
    margin: 0 auto 36px;
  }

  .ve-spotlight__header h2 {
    margin: 0 0 8px;
  }

  .ve-spotlight__header p {
    margin: 0;
    color: rgba(226, 232, 240, 0.7);
  }

  .ve-spotlight__grid {
    width: min(1120px, 90vw);
    margin: 0 auto;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 20px;
  }

  .ve-spotlight__card {
    background: rgba(15, 23, 42, 0.8);
    border: 1px solid rgba(148, 163, 184, 0.2);
    border-radius: 20px;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .ve-spotlight__capacity span {
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.12em;
    color: rgba(226, 232, 240, 0.6);
  }

  .ve-spotlight__capacity strong {
    display: block;
    font-size: 1.3rem;
  }

  .ve-spotlight__card ul {
    list-style: none;
    padding: 0;
    margin: 0;
    display: grid;
    gap: 8px;
    color: rgba(226, 232, 240, 0.7);
  }

  .ve-spotlight__card li::before {
    content: "-";
    color: #38bdf8;
    margin-right: 8px;
  }

  .ve-spotlight__cta {
    color: #38bdf8;
    text-decoration: none;
    font-weight: 600;
    margin-top: auto;
  }

  @media (max-width: 720px) {
    .ve-how__journey {
      padding: 20px;
    }
  }
`;
