/**
 * CompressionCapabilities.ts
 * 
 * Detects WebGL texture compression format support for optimized texture loading.
 * Checks for various compression extensions including DXT/S3TC, ETC, ASTC, and PVRTC.
 */

export interface CompressionFormats {
    s3tc: boolean;           // WEBGL_compressed_texture_s3tc (DXT1, DXT3, DXT5)
    s3tc_srgb: boolean;      // EXT_texture_compression_bptc (BC7)
    etc: boolean;            // WEBGL_compressed_texture_etc (ETC2)
    etc1: boolean;           // WEBGL_compressed_texture_etc1 (ETC1 - deprecated)
    astc: boolean;           // WEBGL_compressed_texture_astc
    pvrtc: boolean;          // WEBGL_compressed_texture_pvrtc (iOS)
    atc: boolean;            // WEBGL_compressed_texture_atc (Adreno)
    bptc: boolean;           // EXT_texture_compression_bptc (BC6H/BC7)
    rgtc: boolean;           // EXT_texture_compression_rgtc (BC4/BC5)
}

export interface CompressionCapabilities {
    formats: CompressionFormats;
    extensions: string[];    // All available compression extension names
    renderer: string;        // GPU renderer info
    vendor: string;          // GPU vendor info
    maxTextureSize: number;  // Maximum texture dimension
}

/**
 * Detects all available texture compression formats supported by the WebGL context.
 * 
 * @param gl - WebGL rendering context
 * @returns CompressionCapabilities object with detailed format support
 */
export function detectCompressionCapabilities(gl: WebGLRenderingContext | WebGL2RenderingContext): CompressionCapabilities {
    const formats: CompressionFormats = {
        s3tc: false,
        s3tc_srgb: false,
        etc: false,
        etc1: false,
        astc: false,
        pvrtc: false,
        atc: false,
        bptc: false,
        rgtc: false,
    };

    const extensions: string[] = [];

    // Check S3TC (DXT1, DXT3, DXT5) - Desktop, widely supported
    const s3tcExt = gl.getExtension('WEBGL_compressed_texture_s3tc') ||
                    gl.getExtension('WEBKIT_WEBGL_compressed_texture_s3tc') ||
                    gl.getExtension('MOZ_WEBGL_compressed_texture_s3tc');
    if (s3tcExt) {
        formats.s3tc = true;
        extensions.push('WEBGL_compressed_texture_s3tc');
    }

    // Check S3TC sRGB variant
    const s3tcSrgbExt = gl.getExtension('WEBGL_compressed_texture_s3tc_srgb');
    if (s3tcSrgbExt) {
        formats.s3tc_srgb = true;
        extensions.push('WEBGL_compressed_texture_s3tc_srgb');
    }

    // Check ETC (ETC2/EAC) - OpenGL ES 3.0+, most Android devices
    const etcExt = gl.getExtension('WEBGL_compressed_texture_etc');
    if (etcExt) {
        formats.etc = true;
        extensions.push('WEBGL_compressed_texture_etc');
    }

    // Check ETC1 (deprecated but still present on some devices)
    const etc1Ext = gl.getExtension('WEBGL_compressed_texture_etc1');
    if (etc1Ext) {
        formats.etc1 = true;
        extensions.push('WEBGL_compressed_texture_etc1');
    }

    // Check ASTC - Modern mobile, excellent quality
    const astcExt = gl.getExtension('WEBGL_compressed_texture_astc');
    if (astcExt) {
        formats.astc = true;
        extensions.push('WEBGL_compressed_texture_astc');
    }

    // Check PVRTC - iOS devices (PowerVR GPUs)
    const pvrtcExt = gl.getExtension('WEBGL_compressed_texture_pvrtc') ||
                     gl.getExtension('WEBKIT_WEBGL_compressed_texture_pvrtc');
    if (pvrtcExt) {
        formats.pvrtc = true;
        extensions.push('WEBGL_compressed_texture_pvrtc');
    }

    // Check ATC - Qualcomm Adreno GPUs (older Android)
    const atcExt = gl.getExtension('WEBGL_compressed_texture_atc');
    if (atcExt) {
        formats.atc = true;
        extensions.push('WEBGL_compressed_texture_atc');
    }

    // Check BPTC (BC6H/BC7) - Desktop, newer
    const bptcExt = gl.getExtension('EXT_texture_compression_bptc');
    if (bptcExt) {
        formats.bptc = true;
        extensions.push('EXT_texture_compression_bptc');
    }

    // Check RGTC (BC4/BC5) - Desktop, for normal maps
    const rgtcExt = gl.getExtension('EXT_texture_compression_rgtc');
    if (rgtcExt) {
        formats.rgtc = true;
        extensions.push('EXT_texture_compression_rgtc');
    }

    // Get GPU info (use standard parameters, WEBGL_debug_renderer_info is deprecated)
    const renderer = gl.getParameter(gl.RENDERER);
    const vendor = gl.getParameter(gl.VENDOR);

    // Get max texture size
    const maxTextureSize = gl.getParameter(gl.MAX_TEXTURE_SIZE);

    return {
        formats,
        extensions,
        renderer: renderer || 'Unknown',
        vendor: vendor || 'Unknown',
        maxTextureSize: maxTextureSize || 0,
    };
}

/**
 * Pretty-print compression capabilities to console.
 * 
 * @param capabilities - Detected compression capabilities
 */
export function logCompressionCapabilities(capabilities: CompressionCapabilities): void {
    console.group('🎨 WebGL Texture Compression Capabilities');
    
    console.log('📊 GPU Information:');
    console.log(`  Vendor: ${capabilities.vendor}`);
    console.log(`  Renderer: ${capabilities.renderer}`);
    console.log(`  Max Texture Size: ${capabilities.maxTextureSize}px`);
    console.log('');

    console.log('📦 Supported Compression Formats:');
    
    // Desktop formats
    if (capabilities.formats.s3tc) {
        console.log('  ✅ S3TC (DXT1/DXT3/DXT5) - Desktop standard');
    } else {
        console.log('  ❌ S3TC (DXT1/DXT3/DXT5) - Not supported');
    }

    if (capabilities.formats.s3tc_srgb) {
        console.log('  ✅ S3TC sRGB - Desktop with sRGB support');
    }

    if (capabilities.formats.bptc) {
        console.log('  ✅ BPTC (BC6H/BC7) - Modern desktop');
    }

    if (capabilities.formats.rgtc) {
        console.log('  ✅ RGTC (BC4/BC5) - Desktop, normal maps');
    }

    // Mobile formats
    if (capabilities.formats.etc) {
        console.log('  ✅ ETC2/EAC - Modern mobile (OpenGL ES 3.0+)');
    } else {
        console.log('  ❌ ETC2/EAC - Not supported');
    }

    if (capabilities.formats.etc1) {
        console.log('  ✅ ETC1 - Legacy mobile');
    }

    if (capabilities.formats.astc) {
        console.log('  ✅ ASTC - High-quality mobile');
    } else {
        console.log('  ❌ ASTC - Not supported');
    }

    if (capabilities.formats.pvrtc) {
        console.log('  ✅ PVRTC - iOS/PowerVR');
    }

    if (capabilities.formats.atc) {
        console.log('  ✅ ATC - Qualcomm Adreno (legacy)');
    }

    console.log('');
    console.log('🔧 Available Extensions:');
    if (capabilities.extensions.length > 0) {
        capabilities.extensions.forEach(ext => console.log(`  • ${ext}`));
    } else {
        console.log('  (No compression extensions available)');
    }

    console.groupEnd();
}

/**
 * Get recommended texture format based on available compression support.
 * 
 * @param capabilities - Detected compression capabilities
 * @returns Recommended format string
 */
export function getRecommendedFormat(capabilities: CompressionCapabilities): string {
    // Priority order: S3TC (desktop) > ASTC (mobile) > ETC2 (mobile) > ETC1 (legacy) > Uncompressed
    if (capabilities.formats.s3tc) {
        return 'S3TC/DXT';
    } else if (capabilities.formats.astc) {
        return 'ASTC';
    } else if (capabilities.formats.etc) {
        return 'ETC2';
    } else if (capabilities.formats.etc1) {
        return 'ETC1';
    } else if (capabilities.formats.pvrtc) {
        return 'PVRTC';
    } else {
        return 'Uncompressed';
    }
}
