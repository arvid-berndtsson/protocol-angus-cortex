#!/usr/bin/env bash

# Script to redistribute existing commits over the last 7 days
# This keeps the actual commit content but changes the dates

echo "Redistributing commits over the last 7 days..."

# Get the number of commits (excluding the initial commit)
COMMIT_COUNT=$(git log --oneline | wc -l)
COMMIT_COUNT=$((COMMIT_COUNT - 1))  # Exclude initial commit

echo "Found $COMMIT_COUNT commits to redistribute"

# Calculate the date range (7 days back from today)
END_DATE=$(date '+%Y-%m-%d')
START_DATE=$(date -v-7d '+%Y-%m-%d')

echo "Date range: $START_DATE to $END_DATE"

# Convert dates to timestamps
START_TIMESTAMP=$(date -j -f "%Y-%m-%d" "$START_DATE" '+%s')
END_TIMESTAMP=$(date -j -f "%Y-%m-%d" "$END_DATE" '+%s')
DATE_RANGE_SECONDS=$((END_TIMESTAMP - START_TIMESTAMP))

# Create a temporary file for the rebase commands
REBASE_FILE=$(mktemp)

# Generate random timestamps for each commit
for ((i=1; i<=COMMIT_COUNT; i++)); do
    # Generate random timestamp within the date range
    RANDOM1=$RANDOM
    RANDOM2=$RANDOM
    RANDOM_COMBINED=$((RANDOM1 * 32768 + RANDOM2))
    RANDOM_SECONDS=$((RANDOM_COMBINED % DATE_RANGE_SECONDS))
    
    # Add some additional randomness
    RANDOM_SECONDS=$((RANDOM_SECONDS + (RANDOM % 3600)))
    
    # Ensure we don't exceed the date range
    if [ $RANDOM_SECONDS -ge $DATE_RANGE_SECONDS ]; then
        RANDOM_SECONDS=$((DATE_RANGE_SECONDS - 1))
    fi
    
    COMMIT_TIMESTAMP=$((START_TIMESTAMP + RANDOM_SECONDS))
    NEW_DATE=$(date -r "$COMMIT_TIMESTAMP" '+%Y-%m-%d %H:%M:%S')
    
    echo "pick $(git log --reverse --oneline | tail -n +2 | sed -n "${i}p" | cut -d' ' -f1) $(git log --reverse --oneline | tail -n +2 | sed -n "${i}p" | cut -d' ' -f2-)"
    echo "exec GIT_AUTHOR_DATE='$NEW_DATE' GIT_COMMITTER_DATE='$NEW_DATE' git commit --amend --no-edit"
    echo ""
done >> "$REBASE_FILE"

echo "Rebase file created. To apply the changes, run:"
echo "git rebase -i --root < $REBASE_FILE"
echo ""
echo "Or to preview the rebase file:"
echo "cat $REBASE_FILE" 