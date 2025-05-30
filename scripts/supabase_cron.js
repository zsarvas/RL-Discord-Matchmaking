// filepath: /C:/Users/zachs/rl_github/RL-Discord-Matchmaking/scripts/supabase_cron.js

const { createClient } = require('@supabase/supabase-js');

const supabaseUrl = process.env.SUPABASE_URL;
const supabaseKey = process.env.SUPABASE_KEY;
const supabase = createClient(supabaseUrl, supabaseKey);

(async () => {
  try {
    // Insert a dummy row
    const { data: insertData, error: insertError } = await supabase
      .from('rocketleague')
      .insert([{ dummy_column: 'dummy_value' }]);

    if (insertError) throw insertError;
    console.log('Inserted row:', insertData);

    // Delete the dummy row
    const { data: deleteData, error: deleteError } = await supabase
      .from('rocketleague')
      .delete()
      .eq('dummy_column', 'dummy_value');

    if (deleteError) throw deleteError;
    console.log('Deleted row:', deleteData);
  } catch (error) {
    console.error('Error:', error.message);
    process.exit(1);
  }
})();