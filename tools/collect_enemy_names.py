
import argparse
import json
from collections import defaultdict
from pathlib import Path
                      
def process_enemy_table(enemy_table: Path, saved_names):
    seen_enemy_index = set([])
    seen_enemy_name = set([])

    with enemy_table.open("r", encoding="utf-8") as f:
        enemy_json = json.load(f)
        enemy_json = enemy_json['EnemyData']
        output_dict = defaultdict(list)
        for key, enemy in enemy_json.items():
            if not key.startswith("enemy_"):
                continue

            if key in saved_names:
                # print(saved_names[key])
                output_dict[key] = saved_names[key]
                continue

            enemyId = enemy['EnemyId']
            enemyIndex = enemy['EnemyIndex']
            name = enemy['Name']

            assert enemyId == key, "{} != {}".format(enemyId, key)
            if enemyIndex in seen_enemy_index:
                print("Duplicate enemy index {},{}".format(key, enemyIndex))
            if name in seen_enemy_name:
                print("Duplicate enemy name {},{}".format(key, name))
            seen_enemy_index.add(enemyIndex)
            seen_enemy_name.add(name)

            name = name.replace("'", "")
            if not name.isascii():
                print(name)

            try:
                if len(output_dict[key]) > 0:
                    output_dict[key].append(name)
                    output_dict[key].append(enemyIndex)
                else:
                    output_dict[key] = [name, enemyIndex]
            except UnicodeEncodeError as e:
                print("UnicodeEncodeError", key, name)
                raise e
            except Exception as e:
                print(e)
                print("Unhandled enemy id: " + key, name) 

        for key in saved_names:
            if key not in output_dict:
                output_dict[key] = saved_names[key]
                
        return output_dict


def main():
    parser = argparse.ArgumentParser(
        description="Collect enemy names from enemy handbook table"
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=Path("./output.json"),
        help="Output filepath (default: ./output.json)"
    )
    args = parser.parse_args()
    
    saved_names = {}
    with Path("./saved_enemy_names.json").open('r', encoding="utf-8") as f:
        saved_names = json.loads(f.read().encode("utf-8"))

    # curl https://raw.githubusercontent.com/Kengxxiao/ArknightsGameData_YoStar/main/en_US/gamedata/excel/enemy_handbook_table.json > enemy_handbook_table.json
    output_dict = process_enemy_table(Path("./enemy_handbook_table.json"), saved_names)

    sorted_keys = sorted(output_dict.keys())
    with args.output.open("w", encoding="utf-8") as f:
        json.dump({key: output_dict[key] for key in sorted_keys}, f, indent=4, ensure_ascii=False)

if __name__ == "__main__":
    main()