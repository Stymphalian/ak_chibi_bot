"""
A tool which compresses textures files in a given directory into compatible webgl
compressed formats.

Usage:
    python tools/compress_textures.py <input_directory> <output_directory> [--format BC3]
    python compress_textures.py ../static/assets/enemies ../static/assets/enemies --format BC2 --ignore-validation
    python compress_textures.py ../static/assets/characters ../static/assets/characters --format BC2 --ignore-validation
When output_directory is the same as input_directory, compressed files are placed
alongside the originals with .dds extension:
    input:  static/assets/characters/char_002_amiya/default/Base/amiya.png
    output: static/assets/characters/char_002_amiya/default/Base/amiya.dds

Steps:
1. Scan the input directory for texture files (e.g., PNG, JPEG).
2. For each texture file, convert it into the specified compressed format.
2.1 Use the compressonator library command (it should be in the thirdparty/compressonatorcli-4.5.52-Linux folder)
    https://github.com/GPUOpen-Tools/compressonator
3. Run in parallel processes to speed up the conversion

Compressed formats available:
- BC1 (DXT1) - RGB, no alpha, 6:1 compression
- BC2 (DXT3) - RGBA, explicit alpha, 4:1 compression
- BC3 (DXT5) - RGBA, interpolated alpha, 4:1 compression (default)

For WEBGL_compressed_texture_s3tc extension.
Mipmaps are NOT generated offline (will be done in renderer).
"""

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

# Available compression formats (format_name, extension, compressonator_format, description)
AVAILABLE_FORMATS = {
    'BC1': ('BC1', '.dds', 'BC1', 'DXT1 - RGB, no alpha, 6:1 compression'),
    'BC2': ('BC2', '.dds', 'BC2', 'DXT3 - RGBA, explicit alpha, 4:1 compression'),
    'BC3': ('BC3', '.dds', 'BC3', 'DXT5 - RGBA, interpolated alpha, 4:1 compression'),
}


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
                    extension: str, same_dir_output: bool = False) -> Path:
    """
    Generate output path maintaining directory structure.
    
    Args:
        input_path: Full path to input file
        input_dir: Root input directory
        output_dir: Root output directory
        extension: File extension for output (e.g., '.dds')
        same_dir_output: If True, place output in same directory as input
        
    Returns:
        Path object for output file
    """
    if same_dir_output:
        # Output in same directory as input, just change extension
        return input_path.with_suffix(extension)
    else:
        # Get relative path from input_dir
        relative_path = input_path.relative_to(input_dir)
        
        # Change extension and maintain directory structure
        output_path = output_dir / relative_path.parent / (relative_path.stem + extension)
        
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


def process_textures(input_dir: Path, output_dir: Path, format_code: str = 'BC3', 
                    ignore_validation: bool = False):
    """
    Process all textures in input directory, generating compressed versions.
    
    Args:
        input_dir: Input directory containing textures
        output_dir: Output directory for compressed textures
        format_code: Compression format to use (BC1, BC2, or BC3)
        ignore_validation: Skip dimension validation if True
    """
    # Validate compressonator CLI exists
    if not COMPRESSONATOR_CLI.exists():
        print(f"Error: Compressonator CLI not found at: {COMPRESSONATOR_CLI}")
        print("Please ensure the tool is installed in thirdparty/compressonatorcli-4.5.52-Linux/")
        sys.exit(1)
    
    # Validate format
    if format_code not in AVAILABLE_FORMATS:
        print(f"Error: Invalid format '{format_code}'")
        print(f"Available formats: {', '.join(AVAILABLE_FORMATS.keys())}")
        sys.exit(1)
    
    format_name, extension, comp_format, description = AVAILABLE_FORMATS[format_code]
    
    # Check if outputting to same directory
    same_dir_output = input_dir == output_dir
    
    print(f"Compression format: {format_name} ({description})")
    if same_dir_output:
        print(f"Output mode: In-place (alongside source files with {extension} extension)")
    else:
        print(f"Output directory: {output_dir}")
    print()
    
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
        if not ignore_validation:
            sys.exit(1)
        else:
            print("Ignoring validation errors as per --ignore-validation flag.")
            print()
    
    if validation_warnings:
        print("⚠️  Warnings (textures will compress but may have compatibility issues):")
        for warning in validation_warnings:
            print(warning)
        print()
    
    print("✓ All textures have valid dimensions")
    print()
    
    # Build list of compression jobs
    jobs = []
    for input_path in texture_files:
        output_path = get_output_path(input_path, input_dir, output_dir, extension, same_dir_output)
        jobs.append((input_path, output_path, comp_format))
    
    # Determine number of processes
    num_processes = max(1, cpu_count() - 1)  # Leave one CPU free
    
    print(f"Using {num_processes} parallel processes")
    print(f"Compressing {len(jobs)} texture(s)...")
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
    if not same_dir_output:
        print(f"Output directory: {output_dir}")
    print("=" * 60)


def main():
    parser = argparse.ArgumentParser(
        description='Compress texture files into WebGL-compatible compressed formats',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Compress textures in-place using BC3 format (default)
  python tools/compress_textures.py static/assets/characters static/assets/characters
  
  # Compress using BC1 format
  python tools/compress_textures.py static/assets/characters static/assets/characters --format BC1
  
  # Compress to separate output directory
  python tools/compress_textures.py static/assets/characters compressed/

Available Formats:
  BC1 (DXT1) - RGB, no alpha, 6:1 compression
  BC2 (DXT3) - RGBA, explicit alpha, 4:1 compression
  BC3 (DXT5) - RGBA, interpolated alpha, 4:1 compression [default]

Output:
  When input and output directories are the same, compressed files are placed
  alongside originals:
    input:  static/assets/characters/char_002/amiya.png
    output: static/assets/characters/char_002/amiya.dds
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
        help='Output directory for compressed textures (can be same as input)'
    )
    
    parser.add_argument(
        '--format',
        type=str,
        default='BC2',
        choices=['BC1', 'BC2', 'BC3'],
        help='Compression format to use (default: BC2/DXT3)'
    )
    
    parser.add_argument(
        "--ignore-validation",
        action="store_true",
        help="Ignore texture dimension validation warnings and errors"
    )
    
    args = parser.parse_args()
    
    # Convert to Path objects
    input_dir = Path(args.input_directory).resolve()
    output_dir = Path(args.output_directory).resolve()
    
    # Validate input directory exists
    if not input_dir.exists():
        print(f"Error: Input directory does not exist: {input_dir}")
        sys.exit(1)
    
    if not output_dir.is_dir():
        print(f"Error: Output path is not a directory: {output_dir}")
        sys.exit(1)
    
    # Create output directory if it doesn't exist (unless same as input)
    if input_dir != output_dir:
        output_dir.mkdir(parents=True, exist_ok=True)
    
    # Process textures
    process_textures(input_dir, output_dir, args.format, ignore_validation=args.ignore_validation)


if __name__ == "__main__":
    main()

