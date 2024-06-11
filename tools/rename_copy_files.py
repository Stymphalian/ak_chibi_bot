import os
import json
from pathlib import Path

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
                    num_renamed += 1
                    new_file = file.replace(num_suffix, "")
                    full_new_path = Path(os.path.join(root, new_file)).absolute()                   
                    print(full_path, full_new_path)
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

rename_copy_files(Path('../server/assets/characters'))
reduce_json_size(Path('../server/assets/characters'))
rename_copy_files(Path('../server/assets/enemies'))
reduce_json_size(Path('../server/assets/enemies'))