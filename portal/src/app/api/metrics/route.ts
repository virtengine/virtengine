/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { NextResponse } from 'next/server';
import { getPriceMetrics } from '@/lib/price-metrics';

export const runtime = 'nodejs';

export async function GET() {
  const metrics = getPriceMetrics();
  const body = await metrics.registry.metrics();

  return new NextResponse(body, {
    headers: {
      'Content-Type': metrics.registry.contentType,
    },
  });
}
