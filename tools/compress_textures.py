"""
A tool which compresses textures files in a given directory into compatible webgl
compressed formats.

Usage:
    python tools/compress_textures.py <input_directory> <output_directory>

Steps:
1. Scan the input directory for texture files (e.g., PNG, JPEG).
2. For each texture file, convert it into compressed formats.
2.1 Use the compressonator library command (it should be in the thirdparty/compressonatorcli-4.5.52-Linux folder)
3. Run in parallel processes to speed up the conversion

Compressed formats generated:
- BC1 (DXT1) - RGB, no alpha, 6:1 compression
- BC2 (DXT3) - RGBA, explicit alpha, 4:1 compression
- BC3 (DXT5) - RGBA, interpolated alpha, 4:1 compression

For WEBGL_compressed_texture_s3tc extension.
Mipmaps are NOT generated offline (will be done in renderer).
"""

import os
import sys
import argparse
import subprocess
from pathlib import Path
from multiprocessing import Pool, cpu_count
from typing import List, Tuple
from PIL import Image


# Path to the Compressonator CLI tool
SCRIPT_DIR = Path(__file__).parent.absolute()
PROJECT_ROOT = SCRIPT_DIR.parent
COMPRESSONATOR_CLI = PROJECT_ROOT / "thirdparty" / "compressonatorcli-4.5.52-Linux" / "compressonatorcli"

# Supported input formats
INPUT_EXTENSIONS = {'.png', '.jpg', '.jpeg', '.tga', '.bmp'}

# Compression formats to generate (format_name, extension, compressonator_format)
COMPRESSION_FORMATS = [
    ('BC1', '.dds', 'BC1'),  # DXT1 - RGB, no alpha, 6:1 compression
    ('BC2', '.dds', 'BC2'),  # DXT3 - RGBA, explicit alpha, 4:1 compression
    ('BC3', '.dds', 'BC3'),  # DXT5 - RGBA, interpolated alpha, 4:1 compression
]


def is_power_of_2(n: int) -> bool:
    """Check if a number is a power of 2."""
    return n > 0 and (n & (n - 1)) == 0


def is_multiple_of_4(n: int) -> bool:
    """Check if a number is a multiple of 4."""
    return n > 0 and n % 4 == 0


def validate_texture_dimensions(image_path: Path) -> Tuple[bool, str, int, int]:
    """
    Validate texture dimensions for DXT compression.
    
    DXT/S3TC formats require dimensions to be multiples of 4.
    Powers of 2 are recommended for best GPU performance.
    
    Args:
        image_path: Path to image file
        
    Returns:
        Tuple of (is_valid: bool, message: str, width: int, height: int)
    """
    try:
        with Image.open(image_path) as img:
            width, height = img.size
            
            # Check if dimensions are multiples of 4 (required for DXT)
            width_valid = is_multiple_of_4(width)
            height_valid = is_multiple_of_4(height)
            
            if not width_valid or not height_valid:
                return (False, 
                       f"Dimensions {width}x{height} not multiples of 4 (required for DXT compression)",
                       width, height)
            
            # Check if dimensions are powers of 2 (recommended)
            width_pot = is_power_of_2(width)
            height_pot = is_power_of_2(height)
            
            if not width_pot or not height_pot:
                return (True,
                       f"⚠️  Dimensions {width}x{height} not powers of 2 (may cause issues on some GPUs)",
                       width, height)
            
            return (True, f"✓ Dimensions {width}x{height} valid", width, height)
            
    except Exception as e:
        return (False, f"Failed to read image: {str(e)}", 0, 0)


def find_texture_files(input_dir: Path) -> List[Path]:
    texture_files = []
    for ext in INPUT_EXTENSIONS:
        texture_files.extend(input_dir.rglob(f'*{ext}'))
    return sorted(texture_files)


def get_output_path(input_path: Path, input_dir: Path, output_dir: Path, 
                    format_name: str, extension: str) -> Path:
    """
    Generate output path maintaining directory structure.
    
    Args:
        input_path: Full path to input file
        input_dir: Root input directory
        output_dir: Root output directory
        format_name: Name of compression format (e.g., 'BC1')
        extension: File extension for output (e.g., '.dds')
        
    Returns:
        Path object for output file
    """
    # Get relative path from input_dir
    relative_path = input_path.relative_to(input_dir)
    
    # Change extension and add format subdirectory
    output_path = output_dir / format_name / relative_path.parent / (relative_path.stem + extension)
    
    return output_path


def compress_texture(args: Tuple[Path, Path, Path]) -> Tuple[bool, str]:
    """
    Compress a single texture file using Compressonator CLI.
    
    Args:
        args: Tuple of (input_path, output_path, format_code)
        
    Returns:
        Tuple of (success: bool, message: str)
    """
    input_path, output_path, format_code = args
    
    # Create output directory if it doesn't exist
    output_path.parent.mkdir(parents=True, exist_ok=True)
    
    # Build compressonator command (use bash to execute the wrapper script)
    cmd = [
        'bash',
        str(COMPRESSONATOR_CLI),
        '-fd', format_code,   # Destination format
        '-nomipmap',          # Do not generate mipmaps (done in renderer)
        str(input_path),      # Source file
        str(output_path),     # Destination file
    ]
    
    try:
        # Run compressonator
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=60  # 60 second timeout per file
        )
        
        if result.returncode == 0:
            return (True, f"✓ {input_path.name} -> {output_path.name}")
        else:
            error_msg = result.stderr.strip() or result.stdout.strip()
            return (False, f"✗ {input_path.name}: {error_msg}")
            
    except subprocess.TimeoutExpired:
        return (False, f"✗ {input_path.name}: Timeout after 60s")
    except Exception as e:
        return (False, f"✗ {input_path.name}: {str(e)}")


def process_textures(input_dir: Path, output_dir: Path):
    """
    Process all textures in input directory, generating compressed versions.
    
    Args:
        input_dir: Input directory containing textures
        output_dir: Output directory for compressed textures
        num_processes: Number of parallel processes (default: CPU count)
    """
    # Validate compressonator CLI exists
    if not COMPRESSONATOR_CLI.exists():
        print(f"Error: Compressonator CLI not found at: {COMPRESSONATOR_CLI}")
        print("Please ensure the tool is installed in thirdparty/compressonatorcli-4.5.52-Linux/")
        sys.exit(1)
    
    # Find all texture files
    print(f"Scanning {input_dir} for texture files...")
    texture_files = find_texture_files(input_dir)
    
    if not texture_files:
        print("No texture files found.")
        return
    
    print(f"Found {len(texture_files)} texture file(s)")
    print()
    
    # Validate texture dimensions
    print("Validating texture dimensions...")
    validation_errors = []
    validation_warnings = []
    
    for texture_path in texture_files:
        is_valid, message, width, height = validate_texture_dimensions(texture_path)
        
        if not is_valid:
            validation_errors.append(f"  ✗ {texture_path.name}: {message}")
        elif "⚠️" in message:
            validation_warnings.append(f"  {message} - {texture_path.name}")
    
    if validation_errors:
        print("❌ Validation failed! The following textures have invalid dimensions:")
        print()
        for error in validation_errors:
            print(error)
        print()
        print("DXT compression requires dimensions to be multiples of 4.")
        print("Please resize these images before compressing.")
        sys.exit(1)
    
    if validation_warnings:
        print("⚠️  Warnings (textures will compress but may have compatibility issues):")
        for warning in validation_warnings:
            print(warning)
        print()
    
    print("✓ All textures have valid dimensions")
    print()
    print(f"Generating {len(COMPRESSION_FORMATS)} format(s) per texture")
    print(f"Total operations: {len(texture_files) * len(COMPRESSION_FORMATS)}")
    print()
    
    # Build list of compression jobs
    jobs = []
    for input_path in texture_files:
        for format_name, extension, format_code in COMPRESSION_FORMATS:
            output_path = get_output_path(input_path, input_dir, output_dir, format_name, extension)
            jobs.append((input_path, output_path, format_code))
    
    # Determine number of processes
    num_processes = max(1, cpu_count() - 1)  # Leave one CPU free
    
    print(f"Using {num_processes} parallel processes")
    print("Compressing textures...")
    print()
    
    # Process in parallel
    success_count = 0
    failure_count = 0
    
    with Pool(processes=num_processes) as pool:
        for success, message in pool.imap_unordered(compress_texture, jobs):
            print(message)
            if success:
                success_count += 1
            else:
                failure_count += 1
    
    print()
    print("=" * 60)
    print(f"Compression complete!")
    print(f"Success: {success_count}")
    print(f"Failed: {failure_count}")
    print(f"Output directory: {output_dir}")
    print("=" * 60)


def main():
    parser = argparse.ArgumentParser(
        description='Compress texture files into WebGL-compatible compressed formats',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Compress all textures in assets/characters to compressed/
  python tools/compress_textures.py static/assets/characters compressed/

Output Structure:
  compressed/
    BC1/       - DXT1 format (.dds) RGB, no alpha
    BC2/       - DXT3 format (.dds) RGBA, explicit alpha
    BC3/       - DXT5 format (.dds) RGBA, interpolated alpha
        """
    )
    
    parser.add_argument(
        'input_directory',
        type=str,
        help='Input directory containing texture files'
    )
    
    parser.add_argument(
        'output_directory',
        type=str,
        help='Output directory for compressed textures'
    )
    
    args = parser.parse_args()
    
    # Convert to Path objects
    input_dir = Path(args.input_directory).resolve()
    output_dir = Path(args.output_directory).resolve()
    
    # Validate input directory exists
    if not input_dir.exists():
        print(f"Error: Input directory does not exist: {input_dir}")
        sys.exit(1)
    
    if not input_dir.is_dir():
        print(f"Error: Input path is not a directory: {input_dir}")
        sys.exit(1)
    
    # Create output directory if it doesn't exist
    output_dir.mkdir(parents=True, exist_ok=True)
    
    # Process textures
    process_textures(input_dir, output_dir)


if __name__ == "__main__":
    main()

