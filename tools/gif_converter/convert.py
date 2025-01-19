import sys
import argparse
from PIL import Image


def convert_to_straightalpha(src):
    # iterate through each pixel in the image and given the alpha value
    # apply pre-multiplied alpha to the pixel and save to a new image
    img_new = Image.new('RGBA', src.size)
    for x in range(src.width):
        for y in range(src.height):
            pixel = src.getpixel((x,y))
            a = pixel[3]
            scaled_a = a / 255
            new_pixel = tuple((
                int(pixel[0] * scaled_a),
                int(pixel[1] * scaled_a),
                int(pixel[2] * scaled_a),
                a
            ))
            img_new.putpixel((x,y), new_pixel)
    return img_new

def create_spritesheet(
        gif_path, 
        columns, 
        flip_left_right=False,
        crop=None,
        convert_to_straight_alpha=False
    ):
    # Open the GIF file
    gif = Image.open(gif_path)
    
    # Calculate the number of frames and the size of each frame
    frames = []
    while True:
        frames.append(gif.copy())
        try:
            gif.seek(len(frames))  # Move to the next frame
        except EOFError:
            break

    # print("Total Number of frames: ", len(frames))
    if crop:
        frames = [frame.crop(crop) for frame in frames]

    frame_width, frame_height = frames[0].size
    rows = (len(frames) + columns - 1) // columns  # Calculate the number of rows needed

    # Create a new image for the spritesheet
    spritesheet = Image.new('RGBA', (columns * frame_width, rows * frame_height))

    # Paste each frame into the spritesheet
    for index, frame in enumerate(frames):
        x = (index % columns) * frame_width
        y = (index // columns) * frame_height
        if flip_left_right:
            frame = frame.transpose(Image.FLIP_LEFT_RIGHT)
        frame = frame.convert('RGBA')
        # print(frame)
        spritesheet.paste(frame, (x, y))

    # Save the spritesheet
    # spritesheet = spritesheet.convert('RGBa')
    if convert_to_straight_alpha:
        spritesheet = convert_to_straightalpha(spritesheet)

    spritesheet = spritesheet.convert("P", palette=Image.ADAPTIVE, colors=256)
    # spritesheet.save("comp.png", optimize=True)

    frame_size = frames[0].size
    spritesheet_size = (rows, columns)
    # print(frame_size, spritesheet_size, len(frames))
    return (spritesheet, frame_size, spritesheet_size, len(frames))


# python main.py --gif_path move.gif
if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Convert GIF to spritesheet')
    parser.add_argument('--gif_path', type=str, help='Path to input GIF file')
    parser.add_argument('--columns', type=int, default=10, help='Number of columns in spritesheet')
    parser.add_argument('--flip_left_right', type=bool, default=False, help='Flip left-right')
    parser.add_argument('--crop', type=str, default=None, help='Crop tuple (left, uppper, right, lower)')
    parser.add_argument('--output', '-o', type=str, help='Output filename (optional)')
    
    args = parser.parse_args()
    gif_path = args.gif_path
    columns = args.columns
    flip_left_right = args.flip_left_right
    crop = args.crop
    if crop is not None:
        crop = tuple(map(int, crop.split(',')))

    output_filename = gif_path.rsplit('.', 1)[0] + '_spritesheet.png'
    # output_filename = 'build_char_10001_solari.move.png'
    spritesheet, _, _, _ = create_spritesheet(gif_path, columns, flip_left_right, crop)
    spritesheet.save(output_filename, 'PNG', optimize=True, compress_level=0)
    print("Finished.")