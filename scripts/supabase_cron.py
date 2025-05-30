from supabase import create_client, Client
import os
import sys

# Load environment variables
SUPABASE_URL = os.getenv("SUPABASE_URL")
SUPABASE_KEY = os.getenv("SUPABASE_KEY")

if not SUPABASE_URL or not SUPABASE_KEY:
    print("Error: Missing Supabase URL or Key in environment variables.")
    sys.exit(1)

# Initialize Supabase client
supabase: Client = create_client(SUPABASE_URL, SUPABASE_KEY)

try:
    # Insert a dummy row
    insert_response = supabase.table("rocketleague").insert({"dummy_column": "dummy_value"}).execute()
    print("Inserted row:", insert_response.data)

    # Delete the dummy row
    delete_response = supabase.table("rocketleague").delete().eq("dummy_column", "dummy_value").execute()
    print("Deleted row:", delete_response.data)

except Exception as e:
    print("Error:", str(e))
    sys.exit(1)