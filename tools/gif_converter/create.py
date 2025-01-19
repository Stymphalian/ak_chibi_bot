import json
import os
import argparse
from collections import defaultdict
from pprint import pprint
from PIL import Image

import convert as gifConvert

"""
Create 
char_10000_bubu
    default
        base
            Front
                build_char_10000_bubu.animations.json
                build_char_10000_bubu.atlas
                build_char_10000_bubu.*.png
                build_char_10000_bubu.spritesheet.json
        battle
            Front
                char_10000_bubu.animations.json
                char_10000_bubu.atlas
                char_10000_bubu.*.png
                char_10000_bubu.spritesheet.json

*.animations.json
    {
    "animations": [
        "Default",
        "Relax",
        "Move",
        "Sit"
    ]
    }

*.atlas    
    build_char_10000_bubu.move.png
    size: 8160,7200
    format: RGBA8888
    filter: Linear,Linear
    repeat: none

    build_char_10000_bubu.relax.png
    size: 8160,7200
    format: RGBA8888
    filter: Linear,Linear
    repeat: none

    build_char_10000_bubu.sit.png
    size: 8160,7200
    format: RGBA8888
    filter: Linear,Linear
    repeat: none

*.spritesheet.json    
{
    "animations": {
        "Relax": {
            "filepath": "build_char_10000_bubu.relax.png",
            "rows": 6,
            "cols": 10, 
            "scaleX": 0.35,
            "scaleY": 0.35,
            "height": 1200,
            "width": 816,
            "frames": 60,
            "fps": 30
        },
        "Move": {
            "filepath": "build_char_10000_bubu.move.png",
            "rows": 6,
            "cols": 10, 
            "scaleX": 0.35,
            "scaleY": 0.35,
            "height": 1200,
            "width": 816,
            "frames": 60,
            "fps": 30
        },
        "Sit": {
            "filepath": "build_char_10000_bubu.sit.png",
            "rows": 6,
            "cols": 10, 
            "scaleX": 0.35,
            "scaleY": 0.35,
            "height": 1200,
            "width": 816,
            "frames": 60,
            "fps": 30
        }
    }
}
"""


"""
char_10000_bubu
    default
        base
            Front
                build_char_10000_bubu.animations.json
                build_char_10000_bubu.atlas
                build_char_10000_bubu.*.png
                build_char_10000_bubu.spritesheet.json
        battle
            Front
                char_10000_bubu.animations.json
                char_10000_bubu.atlas
                char_10000_bubu.*.png
                char_10000_bubu.spritesheet.json
"""
def createFoldersAndFiles(opName):
    # Create the output directory if it doesn't exist
    os.makedirs(opName, exist_ok=True)
    os.makedirs(os.path.join(opName, "default", "base", "Front"), exist_ok=True)
    os.makedirs(os.path.join(opName, "default", "battle", "Front"), exist_ok=True)


# build_char_10000_bubu.animations.json
# build_char_10000_bubu.atlas
# build_char_10000_bubu.*.png
# build_char_10000_bubu.spritesheet.json
def processAnimations(config, outputDir, opName, scales):
    animationJson = {"animations": ["Default"]}
    spritesheetJson = {"animations": {}}
    atlasInfo = defaultdict(tuple)

    for animationName, animation in config["animations"].items():
        print(animationName, animation)

        # create the png file
        gifFile =  animation["filepath"]
        flipLeftRight = animation["flip_left_right"]
        convertAlpha = animation["convert_alpha"]
        crop = animation["crop"] if "crop" in animation else None
        if crop is not None:
            crop = (crop["left"], crop["top"], crop["right"], crop["bottom"])
        spritesheet_png, frame_size_wh, spritsheet_size_row_col, numFrames = gifConvert.create_spritesheet(
            gifFile, 10, flip_left_right=flipLeftRight, crop=crop, convert_to_straight_alpha=convertAlpha)
        png_output_filename = opName + "." + animationName.lower() + ".png"
        png_output_filepath = os.path.join(outputDir, png_output_filename)
        spritesheet_png.save(png_output_filepath, 'PNG', optimize=True, compress_level=0)

        for scaleKey, scale in scales.items():
            scaled_dir = os.path.join(outputDir, scaleKey)
            os.makedirs(scaled_dir, exist_ok=True)
            png_output_filepath = os.path.join(scaled_dir, png_output_filename)
            sheet_size = (int(frame_size_wh[0] * spritsheet_size_row_col[1]), int(frame_size_wh[1] * spritsheet_size_row_col[0]))
            outpng = spritesheet_png.resize((int(sheet_size[0] * scale), int(sheet_size[1] * scale)))
            outpng.save(png_output_filepath, 'PNG', optimize=True, compress_level=0)

        # animations.json file
        animationJson["animations"].append(animationName)
        # create the spritesheet.json file
        spritesheetJson["animations"][animationName] = {
            "filepath": png_output_filename,
            "rows": spritsheet_size_row_col[0],
            "cols": spritsheet_size_row_col[1],
            "scaleX": animation["scaleX"],
            "scaleY": animation["scaleY"],
            "width": frame_size_wh[0],
            "height": frame_size_wh[1],
            "frames": numFrames,
            "fps": animation["fps"]
        }
        # create the atlas file
        atlasInfo[png_output_filename] = (
            frame_size_wh[0] * spritsheet_size_row_col[1],
            frame_size_wh[1] * spritsheet_size_row_col[0]
        )
        

    with open(os.path.join(outputDir, opName + ".animations.json"), 'w') as f:
        json.dump(animationJson, f, indent=4)
    with open(os.path.join(outputDir, opName + ".spritesheet.json"), 'w') as f:
        json.dump(spritesheetJson, f, indent=4)
    with open(os.path.join(outputDir, opName + ".atlas"), 'w') as f:
        f.write("\n")
        for filename, size in atlasInfo.items():
            f.write(filename + "\n")
            f.write("size: " + str(size[0]) + "," + str(size[1]) + "\n")
            f.write("format: RGBA8888\n")
            f.write("filter: Linear,Linear\n")
            f.write("repeat: none\n\n")



# python create.py --op_num 10000 --op_name bubu
def main(args):
    # Example usage
    op_num = args.op_num
    op_name = args.op_name
    fullName = "char_" + str(op_num) + "_" + op_name

    config = None
    with open(args.config_file, 'r') as file:
        config = json.load(file)
    pprint(config)

    createFoldersAndFiles(fullName)
    base_outputdir =  os.path.join(fullName, "default", "base", "Front")
    battle_output_dir = os.path.join(fullName, "default", "battle", "Front")
    processAnimations(config["base"], base_outputdir, "build_" +  fullName, config["scales"])
    processAnimations(config["battle"], battle_output_dir, fullName, config["scales"])
    

if __name__ == "__main__":
    args = argparse.ArgumentParser()
    args.add_argument("--op_num", type=int, help="Operator number. example 10000")  
    args.add_argument("--op_name", type=str, help="Operator name. example bubu")
    args.add_argument("--config_file", type=str, default="spritesheet.json")
    args = args.parse_args()
    main(args)
