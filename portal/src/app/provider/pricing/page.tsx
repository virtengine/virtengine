import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Pricing Configuration',
  description: 'Configure your pricing strategy',
};

export default function ProviderPricingPage() {
  return (
    <div className="container py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Pricing Configuration</h1>
        <p className="mt-1 text-muted-foreground">
          Configure your pricing strategy for compute resources
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Base Pricing */}
        <div className="rounded-lg border border-border bg-card p-6">
          <h2 className="text-lg font-semibold">Base Pricing</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Set your base rates for different resource types
          </p>
          
          <div className="mt-6 space-y-4">
            <PricingInput label="CPU (per core/hour)" defaultValue="0.05" />
            <PricingInput label="Memory (per GB/hour)" defaultValue="0.02" />
            <PricingInput label="GPU (per unit/hour)" defaultValue="1.50" />
            <PricingInput label="Storage (per GB/month)" defaultValue="0.10" />
          </div>
        </div>

        {/* Bid Strategy */}
        <div className="rounded-lg border border-border bg-card p-6">
          <h2 className="text-lg font-semibold">Bid Strategy</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Configure automatic bidding behavior
          </p>

          <div className="mt-6 space-y-4">
            <div>
              <label className="text-sm font-medium" htmlFor="bid-strategy">
                Strategy
              </label>
              <select
                id="bid-strategy"
                className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
              >
                <option>Aggressive - Maximize utilization</option>
                <option>Balanced - Balance price and utilization</option>
                <option>Conservative - Maximize revenue per lease</option>
                <option>Manual - Review all bids manually</option>
              </select>
            </div>

            <div>
              <label className="text-sm font-medium" htmlFor="min-margin">
                Minimum Margin (%)
              </label>
              <input
                type="number"
                id="min-margin"
                defaultValue="15"
                className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
              />
            </div>

            <div>
              <label className="text-sm font-medium" htmlFor="auto-accept">
                Auto-accept threshold ($)
              </label>
              <input
                type="number"
                id="auto-accept"
                defaultValue="100"
                className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
              />
            </div>
          </div>
        </div>

        {/* Volume Discounts */}
        <div className="rounded-lg border border-border bg-card p-6">
          <h2 className="text-lg font-semibold">Volume Discounts</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Offer discounts for longer commitments
          </p>

          <div className="mt-6 space-y-3">
            <DiscountRow duration="1 week" discount="5%" />
            <DiscountRow duration="1 month" discount="10%" />
            <DiscountRow duration="3 months" discount="15%" />
            <DiscountRow duration="1 year" discount="25%" />
          </div>

          <button
            type="button"
            className="mt-4 text-sm text-primary hover:underline"
          >
            + Add custom discount tier
          </button>
        </div>

        {/* Price Alerts */}
        <div className="rounded-lg border border-border bg-card p-6">
          <h2 className="text-lg font-semibold">Price Alerts</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Get notified about market price changes
          </p>

          <div className="mt-6 space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm">Market price drops below base</span>
              <ToggleSwitch defaultChecked />
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm">Competitor price changes</span>
              <ToggleSwitch />
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm">Low utilization warning</span>
              <ToggleSwitch defaultChecked />
            </div>
          </div>
        </div>
      </div>

      {/* Save Button */}
      <div className="mt-8 flex justify-end gap-4">
        <button
          type="button"
          className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent"
        >
          Cancel
        </button>
        <button
          type="button"
          className="rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
        >
          Save Changes
        </button>
      </div>
    </div>
  );
}

function PricingInput({ label, defaultValue }: { label: string; defaultValue: string }) {
  return (
    <div>
      <label className="text-sm font-medium">{label}</label>
      <div className="mt-1 flex items-center gap-2">
        <span className="text-muted-foreground">$</span>
        <input
          type="number"
          defaultValue={defaultValue}
          step="0.01"
          className="flex-1 rounded-lg border border-border bg-background px-3 py-2 text-sm"
        />
      </div>
    </div>
  );
}

function DiscountRow({ duration, discount }: { duration: string; discount: string }) {
  return (
    <div className="flex items-center justify-between rounded-lg border border-border bg-muted/30 p-3">
      <span className="text-sm">{duration}</span>
      <input
        type="text"
        defaultValue={discount}
        className="w-20 rounded border border-border bg-background px-2 py-1 text-right text-sm"
      />
    </div>
  );
}

function ToggleSwitch({ defaultChecked }: { defaultChecked?: boolean }) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={defaultChecked}
      className={`relative h-6 w-11 rounded-full transition-colors ${
        defaultChecked ? 'bg-primary' : 'bg-muted'
      }`}
    >
      <span
        className={`absolute left-0.5 top-0.5 h-5 w-5 rounded-full bg-white transition-transform ${
          defaultChecked ? 'translate-x-5' : ''
        }`}
      />
    </button>
  );
}
