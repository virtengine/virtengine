# Learnings for VirtEngine Blockchain Back-End  

## Technical Insights

<!-- Add useful patterns, solutions, and discoveries here -->

- Encrypted payload envelopes now validate `SenderSignature` and optionally `RecipientPublicKeys`; when public keys are provided they must align with `RecipientKeyIDs` fingerprints to avoid validation failures.

## Blockers & Resolutions

<!-- Document blockers encountered and how they were resolved -->
