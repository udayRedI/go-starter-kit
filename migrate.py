import json
import subprocess
import sys
import re


json_file_path = '/app/system_configs/config.json'


try:
    with open(json_file_path, 'r') as file:
        json_data = json.load(file)

        db_url = json_data.get("DbUrl")

        linux_command = f"atlas migrate apply -u \"{db_url}\" --allow-dirty"
        subprocess.run(linux_command, shell=True, check=True)
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