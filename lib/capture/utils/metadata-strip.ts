/**
 * Metadata Stripping Utility
 * VE-210: Remove EXIF, GPS, and other metadata from captured images
 *
 * Ensures privacy by removing all potentially identifying metadata
 * before encryption and upload.
 */

/**
 * Metadata types that are stripped
 */
export const STRIPPED_METADATA_TYPES = [
  'EXIF',
  'GPS',
  'IPTC',
  'XMP',
  'ICC Profile',
  'Photoshop',
  'Maker Notes',
  'Comments',
] as const;

/**
 * Result of metadata stripping operation
 */
export interface MetadataStripResult {
  /** Cleaned image blob */
  cleanBlob: Blob;
  /** Original size in bytes */
  originalSize: number;
  /** Cleaned size in bytes */
  cleanedSize: number;
  /** Whether metadata was found and removed */
  metadataRemoved: boolean;
  /** Types of metadata that were removed */
  removedTypes: string[];
  /** Processing time in ms */
  processingTimeMs: number;
}

/**
 * JPEG marker constants
 */
const JPEG_MARKERS = {
  SOI: 0xffd8, // Start of Image
  EOI: 0xffd9, // End of Image
  SOS: 0xffda, // Start of Scan
  APP0: 0xffe0, // JFIF
  APP1: 0xffe1, // EXIF/XMP
  APP2: 0xffe2, // ICC Profile
  APP3: 0xffe3,
  APP4: 0xffe4,
  APP5: 0xffe5,
  APP6: 0xffe6,
  APP7: 0xffe7,
  APP8: 0xffe8,
  APP9: 0xffe9,
  APP10: 0xffea,
  APP11: 0xffeb,
  APP12: 0xffec, // Ducky
  APP13: 0xffed, // IPTC/Photoshop
  APP14: 0xffee, // Adobe
  APP15: 0xffef,
  COM: 0xfffe, // Comment
} as const;

/**
 * APP markers to remove (contain metadata)
 */
const MARKERS_TO_REMOVE = new Set([
  JPEG_MARKERS.APP1, // EXIF/XMP
  JPEG_MARKERS.APP2, // ICC Profile
  JPEG_MARKERS.APP3,
  JPEG_MARKERS.APP4,
  JPEG_MARKERS.APP5,
  JPEG_MARKERS.APP6,
  JPEG_MARKERS.APP7,
  JPEG_MARKERS.APP8,
  JPEG_MARKERS.APP9,
  JPEG_MARKERS.APP10,
  JPEG_MARKERS.APP11,
  JPEG_MARKERS.APP12, // Ducky
  JPEG_MARKERS.APP13, // IPTC/Photoshop
  JPEG_MARKERS.APP14, // Adobe
  JPEG_MARKERS.APP15,
  JPEG_MARKERS.COM, // Comments
]);

/**
 * Strip all metadata from an image blob
 *
 * Supports JPEG and PNG formats. For JPEG, removes EXIF, GPS, IPTC, XMP,
 * and other APP markers. For PNG, removes all ancillary chunks except
 * those required for display.
 *
 * @param imageBlob - Original image blob
 * @returns Promise resolving to metadata strip result
 */
export async function stripMetadata(imageBlob: Blob): Promise<MetadataStripResult> {
  const startTime = performance.now();
  const originalSize = imageBlob.size;
  const removedTypes: string[] = [];

  // Read the image data
  const arrayBuffer = await imageBlob.arrayBuffer();
  const data = new Uint8Array(arrayBuffer);

  let cleanedData: Uint8Array;
  let metadataRemoved = false;

  // Detect image type from magic bytes
  if (isJpeg(data)) {
    const result = stripJpegMetadata(data);
    cleanedData = result.data;
    metadataRemoved = result.removed;
    removedTypes.push(...result.types);
  } else if (isPng(data)) {
    const result = stripPngMetadata(data);
    cleanedData = result.data;
    metadataRemoved = result.removed;
    removedTypes.push(...result.types);
  } else {
    // For other formats, use canvas-based stripping
    const result = await stripViaCanvas(imageBlob);
    cleanedData = result.data;
    metadataRemoved = true;
    removedTypes.push('All metadata (canvas re-encode)');
  }

  const cleanBlob = new Blob([cleanedData], { type: imageBlob.type || 'image/jpeg' });
  const processingTimeMs = performance.now() - startTime;

  return {
    cleanBlob,
    originalSize,
    cleanedSize: cleanBlob.size,
    metadataRemoved,
    removedTypes,
    processingTimeMs,
  };
}

/**
 * Check if data is a JPEG image
 */
function isJpeg(data: Uint8Array): boolean {
  return data.length >= 2 && data[0] === 0xff && data[1] === 0xd8;
}

/**
 * Check if data is a PNG image
 */
function isPng(data: Uint8Array): boolean {
  return (
    data.length >= 8 &&
    data[0] === 0x89 &&
    data[1] === 0x50 &&
    data[2] === 0x4e &&
    data[3] === 0x47 &&
    data[4] === 0x0d &&
    data[5] === 0x0a &&
    data[6] === 0x1a &&
    data[7] === 0x0a
  );
}

/**
 * Strip metadata from JPEG data
 */
function stripJpegMetadata(data: Uint8Array): {
  data: Uint8Array;
  removed: boolean;
  types: string[];
} {
  const types: string[] = [];
  const segments: Uint8Array[] = [];
  let removed = false;
  let pos = 0;

  // Verify JPEG SOI marker
  if (data[0] !== 0xff || data[1] !== 0xd8) {
    return { data, removed: false, types: [] };
  }

  // Copy SOI marker
  segments.push(data.slice(0, 2));
  pos = 2;

  while (pos < data.length) {
    // Find next marker
    if (data[pos] !== 0xff) {
      pos++;
      continue;
    }

    // Skip padding bytes
    while (pos < data.length && data[pos] === 0xff) {
      pos++;
    }

    if (pos >= data.length) break;

    const marker = 0xff00 | data[pos];
    pos++;

    // Handle End of Image
    if (marker === JPEG_MARKERS.EOI) {
      segments.push(new Uint8Array([0xff, 0xd9]));
      break;
    }

    // Handle Start of Scan (marks start of image data)
    if (marker === JPEG_MARKERS.SOS) {
      // Include SOS marker and everything after
      segments.push(data.slice(pos - 2));
      break;
    }

    // Markers without length (RST0-RST7, SOI, EOI)
    if ((marker >= 0xffd0 && marker <= 0xffd7) || marker === JPEG_MARKERS.SOI) {
      segments.push(new Uint8Array([0xff, marker & 0xff]));
      continue;
    }

    // Read segment length
    if (pos + 2 > data.length) break;

    const length = (data[pos] << 8) | data[pos + 1];
    if (length < 2 || pos + length > data.length) break;

    // Check if this marker should be removed
    if (MARKERS_TO_REMOVE.has(marker)) {
      removed = true;

      // Identify the type of metadata
      if (marker === JPEG_MARKERS.APP1) {
        const segmentData = data.slice(pos + 2, pos + Math.min(length, 20));
        const str = String.fromCharCode(...segmentData);
        if (str.startsWith('Exif')) {
          types.push('EXIF');
        } else if (str.includes('XMP') || str.includes('http://ns.adobe')) {
          types.push('XMP');
        } else {
          types.push('APP1');
        }
      } else if (marker === JPEG_MARKERS.APP2) {
        types.push('ICC Profile');
      } else if (marker === JPEG_MARKERS.APP12) {
        types.push('Ducky');
      } else if (marker === JPEG_MARKERS.APP13) {
        types.push('IPTC/Photoshop');
      } else if (marker === JPEG_MARKERS.APP14) {
        types.push('Adobe');
      } else if (marker === JPEG_MARKERS.COM) {
        types.push('Comment');
      } else {
        types.push(`APP${marker - 0xffe0}`);
      }

      // Skip this segment
      pos += length;
      continue;
    }

    // Keep this segment (includes APP0/JFIF and required markers)
    segments.push(data.slice(pos - 2, pos + length));
    pos += length;
  }

  // Concatenate all kept segments
  const totalLength = segments.reduce((sum, seg) => sum + seg.length, 0);
  const result = new Uint8Array(totalLength);
  let offset = 0;
  for (const seg of segments) {
    result.set(seg, offset);
    offset += seg.length;
  }

  return { data: result, removed, types: [...new Set(types)] };
}

/**
 * Strip metadata from PNG data
 */
function stripPngMetadata(data: Uint8Array): {
  data: Uint8Array;
  removed: boolean;
  types: string[];
} {
  const types: string[] = [];
  const chunks: Uint8Array[] = [];
  let removed = false;
  let pos = 8; // Skip PNG signature

  // Copy PNG signature
  chunks.push(data.slice(0, 8));

  // Critical chunks that must be kept
  const criticalChunks = new Set(['IHDR', 'PLTE', 'IDAT', 'IEND']);

  // Additional safe chunks to keep (required for display)
  const safeChunks = new Set(['tRNS', 'cHRM', 'gAMA', 'sBIT', 'bKGD', 'pHYs']);

  while (pos + 12 <= data.length) {
    // Read chunk length (4 bytes, big-endian)
    const length =
      (data[pos] << 24) | (data[pos + 1] << 16) | (data[pos + 2] << 8) | data[pos + 3];

    // Read chunk type (4 bytes ASCII)
    const typeBytes = data.slice(pos + 4, pos + 8);
    const chunkType = String.fromCharCode(...typeBytes);

    const chunkEnd = pos + 12 + length; // length + type + data + CRC
    if (chunkEnd > data.length) break;

    // Check if chunk should be kept
    if (criticalChunks.has(chunkType) || safeChunks.has(chunkType)) {
      chunks.push(data.slice(pos, chunkEnd));
    } else {
      removed = true;
      if (chunkType === 'tEXt' || chunkType === 'zTXt' || chunkType === 'iTXt') {
        types.push(`Text (${chunkType})`);
      } else if (chunkType === 'eXIf') {
        types.push('EXIF');
      } else if (chunkType === 'iCCP') {
        types.push('ICC Profile');
      } else if (chunkType === 'tIME') {
        types.push('Timestamp');
      } else {
        types.push(chunkType);
      }
    }

    pos = chunkEnd;

    // Stop at IEND
    if (chunkType === 'IEND') break;
  }

  // Concatenate chunks
  const totalLength = chunks.reduce((sum, chunk) => sum + chunk.length, 0);
  const result = new Uint8Array(totalLength);
  let offset = 0;
  for (const chunk of chunks) {
    result.set(chunk, offset);
    offset += chunk.length;
  }

  return { data: result, removed, types: [...new Set(types)] };
}

/**
 * Strip metadata by re-encoding through canvas
 * This is a fallback for formats we don't handle directly
 */
async function stripViaCanvas(imageBlob: Blob): Promise<{ data: Uint8Array }> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    const url = URL.createObjectURL(imageBlob);

    img.onload = async () => {
      URL.revokeObjectURL(url);

      try {
        const canvas = document.createElement('canvas');
        canvas.width = img.naturalWidth;
        canvas.height = img.naturalHeight;

        const ctx = canvas.getContext('2d');
        if (!ctx) {
          throw new Error('Failed to get canvas context');
        }

        // Draw image (this strips all metadata)
        ctx.drawImage(img, 0, 0);

        // Convert to blob
        const blob = await new Promise<Blob>((resolveBlob, rejectBlob) => {
          canvas.toBlob(
            (b) => {
              if (b) resolveBlob(b);
              else rejectBlob(new Error('Failed to convert canvas to blob'));
            },
            imageBlob.type || 'image/jpeg',
            0.95
          );
        });

        const arrayBuffer = await blob.arrayBuffer();
        resolve({ data: new Uint8Array(arrayBuffer) });
      } catch (err) {
        reject(err);
      }
    };

    img.onerror = () => {
      URL.revokeObjectURL(url);
      reject(new Error('Failed to load image for metadata stripping'));
    };

    img.src = url;
  });
}

/**
 * Verify that an image has no metadata
 * Performs a quick scan for known metadata markers
 */
export function hasMetadata(data: Uint8Array): boolean {
  if (isJpeg(data)) {
    // Scan for APP1-APP15 and COM markers
    for (let i = 0; i < data.length - 1; i++) {
      if (data[i] === 0xff) {
        const marker = 0xff00 | data[i + 1];
        if (MARKERS_TO_REMOVE.has(marker)) {
          return true;
        }
      }
    }
    return false;
  }

  if (isPng(data)) {
    // Scan for metadata chunks
    const metadataChunks = new Set([
      'tEXt',
      'zTXt',
      'iTXt',
      'eXIf',
      'iCCP',
      'tIME',
      'sPLT',
      'hIST',
    ]);
    let pos = 8;

    while (pos + 12 <= data.length) {
      const length =
        (data[pos] << 24) | (data[pos + 1] << 16) | (data[pos + 2] << 8) | data[pos + 3];
      const typeBytes = data.slice(pos + 4, pos + 8);
      const chunkType = String.fromCharCode(...typeBytes);

      if (metadataChunks.has(chunkType)) {
        return true;
      }

      pos += 12 + length;
      if (chunkType === 'IEND') break;
    }
    return false;
  }

  // For unknown formats, assume metadata might be present
  return true;
}
