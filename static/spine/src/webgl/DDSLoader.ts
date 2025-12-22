/**
 * DDSLoader.ts
 * 
 * Loads and parses DDS (DirectDraw Surface) compressed texture files.
 * Supports DXT1 (BC1), DXT3 (BC2), and DXT5 (BC3) formats for use with
 * WEBGL_compressed_texture_s3tc extension.
 */

// DDS file format constants
const DDS_MAGIC = 0x20534444; // "DDS " in ASCII

// DDS Header flags
const DDSD_MIPMAPCOUNT = 0x20000;

// Pixel format flags
const DDPF_FOURCC = 0x4;

// FourCC codes for DXT formats
const FOURCC_DXT1 = 0x31545844; // "DXT1"
const FOURCC_DXT3 = 0x33545844; // "DXT3"
const FOURCC_DXT5 = 0x35545844; // "DXT5"

export interface DDSInfo {
    width: number;
    height: number;
    format: number;          // WebGL compressed format constant
    mipmapCount: number;
    mipmaps: Uint8Array[];   // Array of mipmap data
    isCompressed: boolean;
}

/**
 * Parse a DDS file from an ArrayBuffer.
 * 
 * @param buffer - ArrayBuffer containing DDS file data
 * @returns DDSInfo object with texture information, or null if invalid
 */
export function parseDDS(buffer: ArrayBuffer): DDSInfo | null {
    // Validate buffer size (minimum DDS header is 128 bytes)
    if (!buffer || buffer.byteLength < 128) {
        console.error(`Invalid DDS file: buffer too small (${buffer ? buffer.byteLength : 0} bytes, expected at least 128)`);
        return null;
    }
    
    const header = new Int32Array(buffer, 0, 31);
    
    // Verify DDS magic number
    if (header[0] !== DDS_MAGIC) {
        // Check if this might be an HTML response (common for 404s served as 200)
        const textDecoder = new TextDecoder('utf-8');
        const first256 = new Uint8Array(buffer, 0, Math.min(256, buffer.byteLength));
        const text = textDecoder.decode(first256);
        
        if (text.includes('<html') || text.includes('<!DOCTYPE')) {
            console.error('Invalid DDS file: received HTML response (likely 404 or server error)');
        } else {
            console.error(`Invalid DDS file: wrong magic number (got 0x${header[0].toString(16)}, expected 0x${DDS_MAGIC.toString(16)})`);
        }
        return null;
    }
    
    // Read header fields
    const height = header[3];
    const width = header[4];
    const mipmapCount = (header[2] & DDSD_MIPMAPCOUNT) ? Math.max(1, header[7]) : 1;
    
    // Read pixel format
    const pixelFormatFlags = header[20];
    
    if (!(pixelFormatFlags & DDPF_FOURCC)) {
        console.error('DDS file is not compressed (no FourCC)');
        return null;
    }
    
    const fourCC = header[21];
    
    // Determine format and block size
    let format: number;
    let blockBytes: number;
    let internalFormat: string;
    
    switch (fourCC) {
        case FOURCC_DXT1:
            format = 0x83F0; // COMPRESSED_RGB_S3TC_DXT1_EXT
            blockBytes = 8;
            internalFormat = 'DXT1';
            break;
        case FOURCC_DXT3:
            format = 0x83F2; // COMPRESSED_RGBA_S3TC_DXT3_EXT
            blockBytes = 16;
            internalFormat = 'DXT3';
            break;
        case FOURCC_DXT5:
            format = 0x83F3; // COMPRESSED_RGBA_S3TC_DXT5_EXT
            blockBytes = 16;
            internalFormat = 'DXT5';
            break;
        default:
            console.error(`Unsupported DDS format: ${fourCC.toString(16)}`);
            return null;
    }
    
    // Extract mipmap data
    const mipmaps: Uint8Array[] = [];
    let dataOffset = 128; // DDS header is 128 bytes
    let mipWidth = width;
    let mipHeight = height;
    
    for (let i = 0; i < mipmapCount; i++) {
        const dataLength = Math.max(1, Math.floor((mipWidth + 3) / 4)) *
                          Math.max(1, Math.floor((mipHeight + 3) / 4)) *
                          blockBytes;
        
        const data = new Uint8Array(buffer, dataOffset, dataLength);
        mipmaps.push(data);
        
        dataOffset += dataLength;
        mipWidth = Math.max(1, mipWidth >> 1);
        mipHeight = Math.max(1, mipHeight >> 1);
    }
    
    // console.log(`Loaded DDS texture: ${internalFormat}, ${width}x${height}, ${mipmapCount} mipmap(s)`);
    
    return {
        width,
        height,
        format,
        mipmapCount,
        mipmaps,
        isCompressed: true
    };
}

/**
 * Load a DDS file from a URL.
 * 
 * @param url - URL to DDS file
 * @returns Promise that resolves with DDSInfo or null on error
 */
export function loadDDS(url: string): Promise<DDSInfo | null> {
    return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest()
        // URL-encode special characters like # which have special meaning in URLs
        // Split by / to preserve path structure, encode each segment, then rejoin
        const urlParts = url.split('/');
        const encodedUrl = urlParts.map(part => encodeURIComponent(part)).join('/');
        // Decode the protocol part if present (http:// or https://)
        const finalUrl = encodedUrl.replace(/^http%3A%2F%2F/, 'http://').replace(/^https%3A%2F%2F/, 'https://');
        
        xhr.open('GET', finalUrl, true);
        xhr.responseType = 'arraybuffer';
        
        xhr.onload = () => {
            if (xhr.status === 200) {
                if (!xhr.response || xhr.response.byteLength === 0) {
                    reject(new Error(`DDS file is empty: ${url}`));
                    return;
                }
                const ddsInfo = parseDDS(xhr.response);
                if (!ddsInfo) {
                    reject(new Error(`Failed to parse DDS file: ${url}`));
                    return;
                }
                resolve(ddsInfo);
            } else if (xhr.status === 404) {
                // File not found - this is normal, will fallback to uncompressed
                reject(new Error(`DDS not found: ${url}`));
            } else {
                reject(new Error(`Failed to load DDS: ${xhr.status} ${xhr.statusText} - ${url}`));
            }
        };
        
        xhr.onerror = () => {
            reject(new Error(`Network error loading DDS: ${url}`));
        };
        
        xhr.send();
    });
}