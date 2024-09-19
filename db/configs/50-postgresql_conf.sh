SOURCE_FILE="/opt/postgres/postgresql.conf"
TARGET_FILE="/var/lib/postgresql/data/postgresql.conf"

# Check if the symlink already exists
if [ -L "$TARGET_FILE" ]; then
  echo "Symlink already exists, no action needed."
else
 # Check if the target file already exists
 if [ -e "$TARGET_FILE" ]; then
   # Delete the existing file
   rm "$TARGET_FILE"
   echo "Existing file postgresql.conf deleted."
 fi

 # Create the symlink
 ln -s "$SOURCE_FILE" "$TARGET_FILE"
 echo "Symlink for postgresql.conf created."
fi

