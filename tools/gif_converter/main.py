import sys
from PIL import Image

def create_spritesheet(gif_path, columns, output_filepath):
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

    print("Total Number of frames: ", len(frames))

    frame_width, frame_height = frames[0].size
    rows = (len(frames) + columns - 1) // columns  # Calculate the number of rows needed

    # Create a new image for the spritesheet
    spritesheet = Image.new('RGBA', (columns * frame_width, rows * frame_height))

    # Paste each frame into the spritesheet
    for index, frame in enumerate(frames):
        x = (index % columns) * frame_width
        y = (index // columns) * frame_height
        spritesheet.paste(frame, (x, y))

    # Save the spritesheet
    spritesheet.save(output_filepath, 'PNG')
    print("Finished.")

# python main.py walk.gif 10 
if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python main.py <gif_path> <columns>")
        sys.exit(1)

    gif_path = sys.argv[1]
    columns = int(sys.argv[2])

    output_filename = gif_path.rsplit('.', 1)[0] + '_spritesheet.png'
    create_spritesheet(gif_path, columns, output_filename)