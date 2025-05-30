from supabase import create_client, Client
import os
import sys
import random
import uuid

# Load environment variables
SUPABASE_URL = os.getenv("SUPABASE_URL")
SUPABASE_KEY = os.getenv("SUPABASE_KEY")

if not SUPABASE_URL or not SUPABASE_KEY:
    print("Error: Missing Supabase URL or Key in environment variables.")
    sys.exit(1)

# Initialize Supabase client
supabase: Client = create_client(SUPABASE_URL, SUPABASE_KEY)

try:
    # Generate random values for the columns
    random_id = random.randint(1, 1000000)
    random_name = f"Player{random.randint(1, 1000)}"
    random_mmr = round(random.uniform(0, 3000), 2)
    random_wins = random.randint(0, 100)
    random_losses = random.randint(0, 100)
    random_match_uid = str(uuid.uuid4())
    random_discord_id = random.randint(100000000000000000, 999999999999999999)

    # Insert a row with random values
    insert_response = supabase.table("rocketleague").insert({
        "id": random_id,
        "Name": random_name,
        "MMR": random_mmr,
        "Wins": random_wins,
        "Losses": random_losses,
        "MatchUID": random_match_uid,
        "DiscordId": random_discord_id
    }).execute()

    if insert_response.status_code != 201:  # 201 is the HTTP status code for "Created"
        raise Exception(f"Insert failed: {insert_response.json()}")

    print("Inserted row:", insert_response.data)

    # Delete the inserted row
    delete_response = supabase.table("rocketleague").delete().eq("id", random_id).execute()

    if delete_response.status_code != 200:  # 200 is the HTTP status code for "OK"
        raise Exception(f"Delete failed: {delete_response.json()}")

    print("Deleted row:", delete_response.data)

except Exception as e:
    print("Error:", str(e))
    sys.exit(1)