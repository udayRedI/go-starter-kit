import json
import subprocess
import sys
import re

if len(sys.argv) != 2:
    print("Migration name is mandatory")
    sys.exit(1)

migration_title = sys.argv[1]

json_file_path = 'app/system_configs/config.local.json'


def replace_pattern_in_file(file_path, pattern, replacement):
    try:
        with open(file_path, 'r') as file:
            content = file.read()

        modified_content = re.sub(pattern, replacement, content)

        with open(file_path, 'w') as file:
            file.write(modified_content)

        print(f"Pattern '{pattern}' replaced with '{replacement}' in {file_path}")

    except FileNotFoundError:
        print(f"Error: File '{file_path}' not found.")
        sys.exit(1)
    except Exception as e:
        print(f"An error occurred: {e}")
        sys.exit(1)


try:
    with open(json_file_path, 'r') as file:
        json_data = json.load(file)

        db_url = json_data.get("DbUrl")
        env = json_data.get("ENV")

        replace_pattern_in_file("atlas.hcl", "DB_URL", db_url)

        linux_command = f"atlas migrate diff {migration_title} --dev-url=\"docker://postgres/15/dev?search_path=public\" --to \"file://combined.sql\""
        subprocess.run(linux_command, shell=True, check=True)
        replace_pattern_in_file("atlas.hcl", db_url, "DB_URL")
except subprocess.CalledProcessError as e:
    print(f"Error running the command: {e}")
    sys.exit(1)
except FileNotFoundError:
    print(f"The file {json_file_path} does not exist.")
    sys.exit(1)
except json.JSONDecodeError as e:
    print(f"Error decoding JSON: {e}")
    sys.exit(1)
except Exception as e:
    print(f"An error occurred: {e}")
    sys.exit(1)
