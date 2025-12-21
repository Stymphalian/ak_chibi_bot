#!/usr/bin/env python3
"""
Automated deployment script for StymphalianTestBot assets.

This script automates the following steps:
1. Collect enemy names
2. Collect saved character names
3. Generate index.json files
4. Copy index files to static/assets/
5. Create assets.zip backup

Usage:
    python deploy_assets.py [--skip-zip] [--verbose]
"""

import argparse
import logging
import shutil
import subprocess
import sys
import zipfile
from pathlib import Path


logger = logging.getLogger(__name__)


class DeploymentScript:
    def __init__(self, project_root: Path, skip_zip: bool = False):
        self.project_root = project_root
        self.tools_dir = project_root / "tools"
        self.static_assets_dir = project_root / "static" / "assets"
        self.server_tools_index_dir = project_root / "server" / "tools" / "index"
        self.skip_zip = skip_zip

    def run_command(self, command: list, cwd: Path = None, shell: bool = False) -> bool:
        """Run a shell command and return success status."""
        if cwd is None:
            cwd = self.project_root

        logger.info(
            f"Running: {' '.join(command) if isinstance(command, list) else command}"
        )
        try:
            result = subprocess.run(
                command,
                cwd=cwd,
                shell=shell,
                check=True,
                capture_output=True,
                text=True,
            )
            if result.stdout:
                logger.debug(result.stdout)
            return True
        except subprocess.CalledProcessError as e:
            logger.error(f"Command failed: {e}")
            if e.stderr:
                logger.error(e.stderr)
            return False

    def step_collect_enemy_names(self) -> bool:
        """Step 1: Run collect_enemy_names.py and output directly to assets."""
        logger.info("Step 1: Collecting enemy names...")

        # Output directly to static/assets/saved_enemy_names.json
        current_output_path = self.tools_dir / "saved_enemy_names.json"
        output_path = self.static_assets_dir / "saved_enemy_names.json"

        if not self.run_command(
            [
                sys.executable,
                "collect_enemy_names.py",
                "--output",
                str(current_output_path),
            ],
            cwd=self.tools_dir,
        ):
            return False

        shutil.copy2(current_output_path, output_path)
        logger.info(f"Created {output_path}")
        return True

    def step_collect_saved_names(self) -> bool:
        """Step 2: Run collect_saved_names.py and output directly to assets."""
        logger.info("Step 2: Collecting saved character names...")

        # Output directly to static/assets/saved_names.json
        current_output_path = self.tools_dir / "saved_names.json"
        output_path = self.static_assets_dir / "saved_names.json"

        if not self.run_command(
            [
                sys.executable,
                "collect_saved_names.py",
                "--output",
                str(current_output_path),
            ],
            cwd=self.tools_dir,
        ):
            return False

        shutil.copy2(current_output_path, output_path)
        logger.info(f"Created {output_path}")
        return True

    def step_generate_index_files(self) -> bool:
        """Step 3: Generate index.json files using Go tool."""
        logger.info("Step 4: Generating index.json files...")

        # Run the Go tool
        asset_dir = str((self.project_root / "static" / "assets").resolve())
        output_dir = str(self.server_tools_index_dir.resolve())

        command = [
            "go",
            "run",
            "create_asset_index.go",
            "-assetDir",
            asset_dir,
            "-outputDir",
            output_dir,
        ]

        if not self.run_command(command, cwd=self.server_tools_index_dir):
            return False

        logger.info("Index files generated successfully")
        return True

    def step_copy_index_to_assets(self) -> bool:
        """Step 4: Copy *_index.json files to static/assets/."""
        logger.info("Step 5: Copying index files to static/assets/...")

        index_files = list(self.server_tools_index_dir.glob("*_index.json"))

        if not index_files:
            logger.error("No index files found")
            return False

        for index_file in index_files:
            dest = self.static_assets_dir / index_file.name
            shutil.copy2(index_file, dest)
            logger.info(f"Copied {index_file.name} to {dest}")

        return True

    def step_create_zip(self) -> bool:
        """Step 5: Create assets.zip with required files."""
        if self.skip_zip:
            logger.info("Step 6: Skipping ZIP creation (--skip-zip)")
            return True

        logger.info("Step 6: Creating assets.zip...")

        zip_path = self.project_root / "assets.zip"

        # Files and directories to include
        items_to_zip = [
            "characters/",
            "enemies/",
            "characters_index.json",
            "enemy_index.json",
            "saved_enemy_names.json",
            "saved_names.json",
        ]

        try:
            with zipfile.ZipFile(zip_path, "w", zipfile.ZIP_DEFLATED) as zipf:
                for item in items_to_zip:
                    item_path = self.static_assets_dir / item

                    if not item_path.exists():
                        logger.warning(f"Warning: {item} not found, skipping")
                        continue

                    if item_path.is_dir():
                        # Add directory and all its contents
                        for file_path in item_path.rglob("*"):
                            if file_path.is_file():
                                arcname = str(
                                    file_path.relative_to(self.static_assets_dir)
                                )
                                zipf.write(file_path, arcname)
                                logger.debug(f"  Added: {arcname}")
                    else:
                        # Add single file
                        arcname = item
                        zipf.write(item_path, arcname)
                        logger.debug(f"  Added: {arcname}")

            logger.info(f"Created {zip_path}")
            logger.info(f"ZIP size: {zip_path.stat().st_size / (1024*1024):.2f} MB")
            return True

        except Exception as e:
            logger.error(f"Failed to create ZIP: {e}")
            return False

    def run(self) -> bool:
        """Run all deployment steps."""
        logger.info("=" * 60)
        logger.info("Starting StymphalianTestBot Asset Deployment")
        logger.info("=" * 60)

        steps = [
            ("Collect enemy names", self.step_collect_enemy_names),
            ("Collect saved names", self.step_collect_saved_names),
            ("Generate index files", self.step_generate_index_files),
            ("Copy index to assets", self.step_copy_index_to_assets),
            ("Create assets.zip", self.step_create_zip),
        ]

        for step_name, step_func in steps:
            logger.info("")
            if not step_func():
                logger.error(f"Deployment failed at: {step_name}")
                return False

        logger.info("")
        logger.info("=" * 60)
        logger.info("Deployment completed successfully!")
        logger.info("=" * 60)
        return True


def main():
    parser = argparse.ArgumentParser(
        description="Automate StymphalianTestBot asset deployment"
    )
    parser.add_argument(
        "--skip-zip", action="store_true", help="Skip creating the assets.zip file"
    )
    parser.add_argument(
        "--project-root",
        type=Path,
        default=None,
        help="Project root directory (defaults to script parent directory)",
    )
    parser.add_argument(
        "--verbose",
        "-v",
        action="store_true",
        help="Enable verbose (DEBUG level) logging",
    )

    args = parser.parse_args()

    # Configure logging
    logging.basicConfig(
        level=logging.DEBUG if args.verbose else logging.INFO,
        format="%(asctime)s - %(levelname)s - %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )

    # Determine project root
    if args.project_root:
        project_root = args.project_root
    else:
        # Assume script is in tools/ directory
        script_dir = Path(__file__).parent
        project_root = script_dir.parent

    if not project_root.exists():
        logging.error(f"Project root not found: {project_root}")
        return 1

    # Run deployment
    deployer = DeploymentScript(project_root=project_root, skip_zip=args.skip_zip)

    success = deployer.run()
    return 0 if success else 1


if __name__ == "__main__":
    sys.exit(main())
