
import argparse
import os
import json
from collections import defaultdict
from pathlib import Path

FROM_INTERNAL_CONFIG = True
CHARACTER_KEY = "Characters"
NAME_KEY = "Name" if FROM_INTERNAL_CONFIG else "name"
                      
def process_character_table(character_table_path: Path, saved_names):
    existing_saved_names = set(saved_names.keys())
    output_dict = defaultdict(list)
    with character_table_path.open("r", encoding="utf-8") as f:
        character_json = json.load(f)
        character_json = character_json[CHARACTER_KEY] if FROM_INTERNAL_CONFIG else character_json
        
        for key, operator in character_json.items():
            if not key.startswith("char_"):
                continue

            if key in saved_names:
                # print(saved_names[key])
                output_dict[key] = saved_names[key]
                existing_saved_names.remove(key)
                continue

            try:
                if key[len(key)-1].isdigit():
                    raise Exception("Unhandled alter operator id: " + key)
                output_dict[key] = [operator[NAME_KEY]]

                # if len(output_dict[key].split(" ")) > 1:
                #     raise Exception("Unhandled multi word Name: ", key, output_dict[key])
            except UnicodeEncodeError as e:
                print("UnicodeEncodeError", key, operator[NAME_KEY])
                raise e
            except:
                print("Unhandled operator id: " + key, operator[NAME_KEY])
                # output_dict[operator_id] = (unicodedata
                #     .normalize("NFKD", operator["name"])
                #     .encode('ascii', "ignore")
                #     .decode())  

    for key in existing_saved_names:
        output_dict[key] = saved_names[key]
    return output_dict


def main():
    parser = argparse.ArgumentParser(
        description="Collect saved character names from character table"
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=None,
        help="Output filepath (default: ./output.json)"
    )
    args = parser.parse_args()
    
    debug=False
    saved_names = {}
    
    currentdir = Path(os.getcwd())
    if debug:
        currentdir = currentdir / "tools"

    saved_names_path = currentdir / Path("saved_names.json")
    character_table_path = currentdir / Path("character_table.json")
    output_path = args.output if args.output else currentdir / Path("output.json")

    with saved_names_path.open(encoding="utf-8") as f:
        saved_names = json.loads(f.read().encode("utf-8"))

    output_dict = process_character_table(character_table_path, saved_names)

    sorted_keys = sorted(output_dict.keys())
    with output_path.open("w", encoding="utf-8") as f:
        json.dump({k: output_dict[k] for k in sorted_keys}, f, indent=4, ensure_ascii=False)

if __name__ == "__main__":
    main()