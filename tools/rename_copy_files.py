import os
import json
from pathlib import Path

def files_binary_equal(full_path, full_new_path):
    with open(full_path, 'rb') as f1, open(full_new_path, 'rb') as f2:
        return f1.read() == f2.read()

def rename_copy_files(directory):
    num_renamed = 0
    num_of_files = 0
    for root, dirs, files in os.walk(directory):
        for file in files:
            num_of_files += 1
            full_path = Path(os.path.join(root, file)).absolute()

            for num in range(5):
                num_suffix = f" ({num})"
                if num_suffix in file:                    
                    new_file = file.replace(num_suffix, "")
                    full_new_path = Path(os.path.join(root, new_file)).absolute()
                    
                    if full_new_path.exists():
                        if not files_binary_equal(full_path, full_new_path):
                            # If the destination path already exists, check if
                            # the file is just a duplicate. If it is then just skip
                            # else we should throw an error
                            print(f"\nFile already exists: {full_new_path}")
                        continue

                    print(full_path, full_new_path)
                    num_renamed += 1
                    os.rename(full_path, full_new_path)

    print("Renamed {}/{} files".format(num_renamed, num_of_files))
            
def reduce_json_size(directory):
    num_updated = 0
    num_of_files = 0
    for root, dirs, files in os.walk(directory):
        for file in files:
            num_of_files += 1
            full_path = Path(os.path.join(root, file)).absolute()
            if ".skel.json" in file:
                num_updated += 1
                
                with open(full_path, 'r') as f:
                    data = json.load(f)
                
                new_data = {
                    'animations': [k for k in data['animations'].keys()]
                }

                new_file = file.replace(".skel.json", ".animations.json")
                full_new_path = Path(os.path.join(root, new_file)).absolute()
                with open(full_new_path, 'w') as f:
                    json.dump(new_data, f, indent=2)

                full_path.unlink()

    print("Udpated json for {}/{} files".format(num_updated, num_of_files))

rename_copy_files(Path('../static/assets/characters'))
reduce_json_size(Path('../static/assets/characters'))
rename_copy_files(Path('../static/assets/enemies'))
reduce_json_size(Path('../static/assets/enemies'))