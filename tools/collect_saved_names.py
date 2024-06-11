
import json
from collections import defaultdict
from pathlib import Path
                      
def process_character_table(character_table_path: Path, saved_names):
    with character_table_path.open("r", encoding="utf-8") as f:
        character_json = json.load(f)
        output_dict = defaultdict(list)
        for key, operator in character_json.items():
            if not key.startswith("char_"):
                continue

            if key in saved_names:
                print(saved_names[key])
                output_dict[key] = saved_names[key]
                continue

            try:
                if key[len(key)-1].isdigit():
                    raise Exception("Unhandled alter operator id: " + key)
                output_dict[key] = [operator["name"]]

                # if len(output_dict[key].split(" ")) > 1:
                #     raise Exception("Unhandled multi word name: ", key, output_dict[key])
            except UnicodeEncodeError as e:
                print("UnicodeEncodeError", key, operator["name"])
                raise e
            except:
                print("Unhandled operator id: " + key, operator["name"])
                # output_dict[operator_id] = (unicodedata
                #     .normalize("NFKD", operator["name"])
                #     .encode('ascii', "ignore")
                #     .decode())            
        return output_dict


def main():
    saved_names = {}
    with Path("./saved_names.json").open(encoding="utf-8") as f:
        saved_names = json.loads(f.read().encode("utf-8"))

    output_dict = process_character_table(Path("./character_table.json"), saved_names)
    with Path("./output.json").open("w", encoding="utf-8") as f:
        json.dump(output_dict, f, indent=4, ensure_ascii=False)

if __name__ == "__main__":
    main()