WITH matches_players AS (
    SELECT
    mmtpp.match_id,
    (
        SELECT DISTINCT player_id FROM matches_maps_teams_players_performance
        WHERE match_id = mmtpp.match_id
        ORDER BY team_id
        LIMIT 1
    ) AS team_1_player_1_id,
    (
        SELECT DISTINCT player_id FROM matches_maps_teams_players_performance
        WHERE match_id = mmtpp.match_id
        ORDER BY team_id
        LIMIT 1 OFFSET 1 
    ) AS team_1_player_2_id,
    (
        SELECT DISTINCT player_id FROM matches_maps_teams_players_performance
        WHERE match_id = mmtpp.match_id
        ORDER BY team_id
        LIMIT 1 OFFSET 2 
    ) AS team_1_player_3_id,
    (
        SELECT DISTINCT player_id FROM matches_maps_teams_players_performance
        WHERE match_id = mmtpp.match_id
        ORDER BY team_id
        LIMIT 1 OFFSET 3 
    ) AS team_1_player_4_id,
    (
        SELECT DISTINCT player_id FROM matches_maps_teams_players_performance
        WHERE match_id = mmtpp.match_id
        ORDER BY team_id
        LIMIT 1 OFFSET 4 
    ) AS team_1_player_5_id,
    (
        SELECT DISTINCT player_id FROM matches_maps_teams_players_performance
        WHERE match_id = mmtpp.match_id
        ORDER BY team_id
        LIMIT 1 OFFSET 5 
    ) AS team_2_player_1_id,
    (
        SELECT DISTINCT player_id FROM matches_maps_teams_players_performance
        WHERE match_id = mmtpp.match_id
        ORDER BY team_id
        LIMIT 1 OFFSET 6 
    ) AS team_2_player_2_id,
    (
        SELECT DISTINCT player_id FROM matches_maps_teams_players_performance
        WHERE match_id = mmtpp.match_id
        ORDER BY team_id
        LIMIT 1 OFFSET 7 
    ) AS team_2_player_3_id,
    (
        SELECT DISTINCT player_id FROM matches_maps_teams_players_performance
        WHERE match_id = mmtpp.match_id
        ORDER BY team_id
        LIMIT 1 OFFSET 8 
    ) AS team_2_player_4_id,
    (
        SELECT DISTINCT player_id FROM matches_maps_teams_players_performance
        WHERE match_id = mmtpp.match_id
        ORDER BY team_id
        LIMIT 1 OFFSET 9 
    ) AS team_2_player_5_id
    FROM matches_maps_teams_players_performance AS mmtpp
    GROUP BY mmtpp.match_id
),
mmtpp_with_date AS (
    SELECT mmtpp.*, matches.date FROM matches_maps_teams_players_performance as mmtpp
    JOIN matches ON mmtpp.match_id = matches.id
)

SELECT 
    m.id,
    m.team_1_id,
    m.team_2_id,
    m.date,
    -- NOTE: Team 1 win rate
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches m2
                WHERE (m2.team_1_id = m.team_1_id OR m2.team_2_id = m.team_1_id)
                AND m2.date < m.date
                AND m2.team_won = m.team_1_id
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches m2
                    WHERE (m2.team_1_id = m.team_1_id OR m2.team_2_id = m.team_1_id)
                    AND m2.date < m.date
                ), 0
                ), 2
        ),0) AS team_1_win_rate,
    -- NOTE: Team 2 win rate
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches m2
                WHERE (m2.team_1_id = m.team_2_id OR m2.team_2_id = m.team_2_id)
                AND m2.date < m.date
                AND m2.team_won = m.team_2_id
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches m2
                    WHERE (m2.team_1_id = m.team_2_id OR m2.team_2_id = m.team_2_id)
                    AND m2.date < m.date
                ), 0
                ), 2
        ),0) AS team_2_win_rate,
    -- NOTE: Team 1 played matches
    (
        SELECT COUNT(*)
        FROM matches
        WHERE (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
        AND matches.date < m.date 
    ) AS team_1_match_played,
    (
    -- NOTE: Team 2 played matches
        SELECT COUNT(*)
        FROM matches
        WHERE (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
        AND matches.date < m.date 
    ) AS team_2_match_played,
    -- NOTE: Team 1 played times on Ascent
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Ascent' LIMIT 1)    
    ) AS team_1_played_times_on_Ascent,
    -- NOTE: Team 1 win rate on Ascent
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_1_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_1_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Ascent' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Ascent' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_1_win_rate_on_Ascent,
    -- NOTE: Team 2 played times on Ascent
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Ascent' LIMIT 1)    
    ) AS team_2_played_times_on_Ascent,
    -- NOTE: Team 2 win rate on Ascent
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_2_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_2_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Ascent' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Ascent' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_2_win_rate_on_Ascent,
    -- NOTE: Team 1 played times on Bind
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Bind' LIMIT 1)    
    ) AS team_1_played_times_on_Bind,
    -- NOTE: Team 1 win rate on Bind
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_1_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_1_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Bind' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Bind' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_1_win_rate_on_Bind,
    -- NOTE: Team 2 played times on Bind
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Bind' LIMIT 1)    
    ) AS team_2_played_times_on_Bind,
    -- NOTE: Team 2 win rate on Bind
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_2_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_2_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Bind' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Bind' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_2_win_rate_on_Bind,
    -- NOTE: Team 1 played times on Breeze
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Breeze' LIMIT 1)    
    ) AS team_1_played_times_on_Breeze,
    -- NOTE: Team 1 win rate on Breeze
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_1_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_1_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Breeze' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Breeze' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_1_win_rate_on_Breeze,
    -- NOTE: Team 2 played times on Breeze
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Breeze' LIMIT 1)    
    ) AS team_2_played_times_on_Breeze,
    -- NOTE: Team 2 win rate on Breeze
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_2_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_2_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Breeze' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Breeze' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_2_win_rate_on_Breeze,
    -- NOTE: Team 1 played times on Haven
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Haven' LIMIT 1)    
    ) AS team_1_played_times_on_Haven,
    -- NOTE: Team 1 win rate on Haven
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_1_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_1_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Haven' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Haven' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_1_win_rate_on_Haven,
    -- NOTE: Team 2 played times on Haven
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Haven' LIMIT 1)    
    ) AS team_2_played_times_on_Haven,
    -- NOTE: Team 2 win rate on Haven
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_2_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_2_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Haven' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Haven' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_2_win_rate_on_Haven,
    -- NOTE: Team 1 played times on Icebox
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Icebox' LIMIT 1)    
    ) AS team_1_played_times_on_Icebox,
    -- NOTE: Team 1 win rate on Icebox
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_1_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_1_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Icebox' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Icebox' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_1_win_rate_on_Icebox,
    -- NOTE: Team 2 played times on Icebox
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Icebox' LIMIT 1)    
    ) AS team_2_played_times_on_Icebox,
    -- NOTE: Team 2 win rate on Icebox
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_2_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_2_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Icebox' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Icebox' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_2_win_rate_on_Icebox,
    -- NOTE: Team 1 played times on Lotus
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Lotus' LIMIT 1)    
    ) AS team_1_played_times_on_Lotus,
    -- NOTE: Team 1 win rate on Lotus
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_1_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_1_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Lotus' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Lotus' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_1_win_rate_on_Lotus,
    -- NOTE: Team 2 played times on Lotus
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Lotus' LIMIT 1)    
    ) AS team_2_played_times_on_Lotus,
    -- NOTE: Team 2 win rate on Lotus
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_2_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_2_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Lotus' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Lotus' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_2_win_rate_on_Lotus,
    -- NOTE: Team 1 played times on Pearl
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Pearl' LIMIT 1)    
    ) AS team_1_played_times_on_Pearl,
    -- NOTE: Team 1 win rate on Pearl
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_1_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_1_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Pearl' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Pearl' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_1_win_rate_on_Pearl,
    -- NOTE: Team 2 played times on Pearl
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Pearl' LIMIT 1)    
    ) AS team_2_played_times_on_Pearl,
    -- NOTE: Team 2 win rate on Pearl
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_2_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_2_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Pearl' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Pearl' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_2_win_rate_on_Pearl,
    -- NOTE: Team 1 played times on Split
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Split' LIMIT 1)    
    ) AS team_1_played_times_on_Split,
    -- NOTE: Team 1 win rate on Split
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_1_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_1_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Split' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Split' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_1_win_rate_on_Split,
    -- NOTE: Team 2 played times on Split
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Split' LIMIT 1)    
    ) AS team_2_played_times_on_Split,
    -- NOTE: Team 2 win rate on Split
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_2_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_2_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Split' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Split' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_2_win_rate_on_Split,
    -- NOTE: Team 1 played times on Sunset
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Sunset' LIMIT 1)    
    ) AS team_1_played_times_on_Sunset,
    -- NOTE: Team 1 win rate on Sunset
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_1_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_1_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Sunset' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_1_id OR matches.team_2_id = m.team_1_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Sunset' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_1_win_rate_on_Sunset,
    -- NOTE: Team 2 played times on Sunset
    (
        SELECT COUNT(*)
        FROM matches 
        FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
        WHERE 
        (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
        AND matches.date < m.date
        AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Sunset' LIMIT 1)    
    ) AS team_2_played_times_on_Sunset,
    -- NOTE: Team 2 win rate on Sunset
    COALESCE( ROUND(
            (
                SELECT COUNT(*)
                FROM matches 
                FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                WHERE (
                    (matches.team_1_id = m.team_2_id AND matches_maps.team_1_score > matches_maps.team_2_score) 
                    OR 
                    (matches.team_2_id = m.team_2_id AND matches_maps.team_2_score > matches_maps.team_1_score)
                )
                AND matches.date < m.date
                AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Sunset' LIMIT 1)    
                ) * 1.0 / NULLIF(
                (
                    SELECT COUNT(*)
                    FROM matches 
                    FULL OUTER JOIN matches_maps ON matches_maps.match_id = matches.id
                    WHERE 
                    (matches.team_1_id = m.team_2_id OR matches.team_2_id = m.team_2_id)
                    AND matches.date < m.date
                    AND matches_maps.map_id = (SELECT id FROM maps WHERE maps.name='Sunset' LIMIT 1)    
                ), 0
                ), 2
        ),0) AS team_2_win_rate_on_Sunset,
    -- NOTE: Statistic for team_1_player_1
    (SELECT AVG(mmtpp_with_date.rating) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_1_id AND mmtpp_with_date.date < m.date) AS team_1_player_1_avg_rating,
    (SELECT AVG(mmtpp_with_date.acs) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_1_id AND mmtpp_with_date.date < m.date) AS team_1_player_1_avg_acs,
    (SELECT AVG(mmtpp_with_date.kast) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_1_id AND mmtpp_with_date.date < m.date) AS team_1_player_1_avg_kast,
    (SELECT AVG(mmtpp_with_date.adr) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_1_id AND mmtpp_with_date.date < m.date) AS team_1_player_1_avg_adr,
    (SELECT AVG(mmtpp_with_date.first_kills) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_1_id AND mmtpp_with_date.date < m.date) AS team_1_player_1_avg_first_kills,
    (SELECT AVG(mmtpp_with_date.first_deaths) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_1_id AND mmtpp_with_date.date < m.date) AS team_1_player_1_avg_first_deaths,
    (SELECT AVG(mmtpp_with_date."2k" + mmtpp_with_date."3k" + mmtpp_with_date."4k" + mmtpp_with_date."5k") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_1_id AND mmtpp_with_date.date < m.date) AS team_1_player_1_avg_multikills,
    (SELECT AVG(mmtpp_with_date."1v1" + mmtpp_with_date."1v2" + mmtpp_with_date."1v3" + mmtpp_with_date."1v4" + mmtpp_with_date."1v5") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_1_id AND mmtpp_with_date.date < m.date) AS team_1_player_1_avg_clutches,
    -- NOTE: Statistic for team_1_player_2
    (SELECT AVG(mmtpp_with_date.rating) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_2_id AND mmtpp_with_date.date < m.date) AS team_1_player_2_avg_rating,
    (SELECT AVG(mmtpp_with_date.acs) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_2_id AND mmtpp_with_date.date < m.date) AS team_1_player_2_avg_acs,
    (SELECT AVG(mmtpp_with_date.kast) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_2_id AND mmtpp_with_date.date < m.date) AS team_1_player_2_avg_kast,
    (SELECT AVG(mmtpp_with_date.adr) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_2_id AND mmtpp_with_date.date < m.date) AS team_1_player_2_avg_adr,
    (SELECT AVG(mmtpp_with_date.first_kills) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_2_id AND mmtpp_with_date.date < m.date) AS team_1_player_2_avg_first_kills,
    (SELECT AVG(mmtpp_with_date.first_deaths) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_2_id AND mmtpp_with_date.date < m.date) AS team_1_player_2_avg_first_deaths,
    (SELECT AVG(mmtpp_with_date."2k" + mmtpp_with_date."3k" + mmtpp_with_date."4k" + mmtpp_with_date."5k") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_2_id AND mmtpp_with_date.date < m.date) AS team_1_player_2_avg_multikills,
    (SELECT AVG(mmtpp_with_date."1v1" + mmtpp_with_date."1v2" + mmtpp_with_date."1v3" + mmtpp_with_date."1v4" + mmtpp_with_date."1v5") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_2_id AND mmtpp_with_date.date < m.date) AS team_1_player_2_avg_clutches,
    -- NOTE: Statistic for team_1_player_3
    (SELECT AVG(mmtpp_with_date.rating) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_3_id AND mmtpp_with_date.date < m.date) AS team_1_player_3_avg_rating,
    (SELECT AVG(mmtpp_with_date.acs) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_3_id AND mmtpp_with_date.date < m.date) AS team_1_player_3_avg_acs,
    (SELECT AVG(mmtpp_with_date.kast) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_3_id AND mmtpp_with_date.date < m.date) AS team_1_player_3_avg_kast,
    (SELECT AVG(mmtpp_with_date.adr) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_3_id AND mmtpp_with_date.date < m.date) AS team_1_player_3_avg_adr,
    (SELECT AVG(mmtpp_with_date.first_kills) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_3_id AND mmtpp_with_date.date < m.date) AS team_1_player_3_avg_first_kills,
    (SELECT AVG(mmtpp_with_date.first_deaths) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_3_id AND mmtpp_with_date.date < m.date) AS team_1_player_3_avg_first_deaths,
    (SELECT AVG(mmtpp_with_date."2k" + mmtpp_with_date."3k" + mmtpp_with_date."4k" + mmtpp_with_date."5k") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_3_id AND mmtpp_with_date.date < m.date) AS team_1_player_3_avg_multikills,
    (SELECT AVG(mmtpp_with_date."1v1" + mmtpp_with_date."1v2" + mmtpp_with_date."1v3" + mmtpp_with_date."1v4" + mmtpp_with_date."1v5") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_3_id AND mmtpp_with_date.date < m.date) AS team_1_player_3_avg_clutches,
    -- NOTE: Statistic for team_1_player_4
    (SELECT AVG(mmtpp_with_date.rating) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_4_id AND mmtpp_with_date.date < m.date) AS team_1_player_4_avg_rating,
    (SELECT AVG(mmtpp_with_date.acs) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_4_id AND mmtpp_with_date.date < m.date) AS team_1_player_4_avg_acs,
    (SELECT AVG(mmtpp_with_date.kast) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_4_id AND mmtpp_with_date.date < m.date) AS team_1_player_4_avg_kast,
    (SELECT AVG(mmtpp_with_date.adr) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_4_id AND mmtpp_with_date.date < m.date) AS team_1_player_4_avg_adr,
    (SELECT AVG(mmtpp_with_date.first_kills) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_4_id AND mmtpp_with_date.date < m.date) AS team_1_player_4_avg_first_kills,
    (SELECT AVG(mmtpp_with_date.first_deaths) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_4_id AND mmtpp_with_date.date < m.date) AS team_1_player_4_avg_first_deaths,
    (SELECT AVG(mmtpp_with_date."2k" + mmtpp_with_date."3k" + mmtpp_with_date."4k" + mmtpp_with_date."5k") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_4_id AND mmtpp_with_date.date < m.date) AS team_1_player_4_avg_multikills,
    (SELECT AVG(mmtpp_with_date."1v1" + mmtpp_with_date."1v2" + mmtpp_with_date."1v3" + mmtpp_with_date."1v4" + mmtpp_with_date."1v5") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_4_id AND mmtpp_with_date.date < m.date) AS team_1_player_4_avg_clutches,
    -- NOTE: Statistic for team_1_player_5
    (SELECT AVG(mmtpp_with_date.rating) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_5_id AND mmtpp_with_date.date < m.date) AS team_1_player_5_avg_rating,
    (SELECT AVG(mmtpp_with_date.acs) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_5_id AND mmtpp_with_date.date < m.date) AS team_1_player_5_avg_acs,
    (SELECT AVG(mmtpp_with_date.kast) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_5_id AND mmtpp_with_date.date < m.date) AS team_1_player_5_avg_kast,
    (SELECT AVG(mmtpp_with_date.adr) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_5_id AND mmtpp_with_date.date < m.date) AS team_1_player_5_avg_adr,
    (SELECT AVG(mmtpp_with_date.first_kills) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_5_id AND mmtpp_with_date.date < m.date) AS team_1_player_5_avg_first_kills,
    (SELECT AVG(mmtpp_with_date.first_deaths) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_5_id AND mmtpp_with_date.date < m.date) AS team_1_player_5_avg_first_deaths,
    (SELECT AVG(mmtpp_with_date."2k" + mmtpp_with_date."3k" + mmtpp_with_date."4k" + mmtpp_with_date."5k") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_5_id AND mmtpp_with_date.date < m.date) AS team_1_player_5_avg_multikills,
    (SELECT AVG(mmtpp_with_date."1v1" + mmtpp_with_date."1v2" + mmtpp_with_date."1v3" + mmtpp_with_date."1v4" + mmtpp_with_date."1v5") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_1_player_5_id AND mmtpp_with_date.date < m.date) AS team_1_player_5_avg_clutches,
    -- NOTE: Statistic for team_2_player_1
    (SELECT AVG(mmtpp_with_date.rating) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_1_id AND mmtpp_with_date.date < m.date) AS team_2_player_1_avg_rating,
    (SELECT AVG(mmtpp_with_date.acs) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_1_id AND mmtpp_with_date.date < m.date) AS team_2_player_1_avg_acs,
    (SELECT AVG(mmtpp_with_date.kast) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_1_id AND mmtpp_with_date.date < m.date) AS team_2_player_1_avg_kast,
    (SELECT AVG(mmtpp_with_date.adr) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_1_id AND mmtpp_with_date.date < m.date) AS team_2_player_1_avg_adr,
    (SELECT AVG(mmtpp_with_date.first_kills) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_1_id AND mmtpp_with_date.date < m.date) AS team_2_player_1_avg_first_kills,
    (SELECT AVG(mmtpp_with_date.first_deaths) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_1_id AND mmtpp_with_date.date < m.date) AS team_2_player_1_avg_first_deaths,
    (SELECT AVG(mmtpp_with_date."2k" + mmtpp_with_date."3k" + mmtpp_with_date."4k" + mmtpp_with_date."5k") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_1_id AND mmtpp_with_date.date < m.date) AS team_2_player_1_avg_multikills,
    (SELECT AVG(mmtpp_with_date."1v1" + mmtpp_with_date."1v2" + mmtpp_with_date."1v3" + mmtpp_with_date."1v4" + mmtpp_with_date."1v5") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_1_id AND mmtpp_with_date.date < m.date) AS team_2_player_1_avg_clutches,
    -- NOTE: Statistic for team_2_player_2
    (SELECT AVG(mmtpp_with_date.rating) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_2_id AND mmtpp_with_date.date < m.date) AS team_2_player_2_avg_rating,
    (SELECT AVG(mmtpp_with_date.acs) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_2_id AND mmtpp_with_date.date < m.date) AS team_2_player_2_avg_acs,
    (SELECT AVG(mmtpp_with_date.kast) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_2_id AND mmtpp_with_date.date < m.date) AS team_2_player_2_avg_kast,
    (SELECT AVG(mmtpp_with_date.adr) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_2_id AND mmtpp_with_date.date < m.date) AS team_2_player_2_avg_adr,
    (SELECT AVG(mmtpp_with_date.first_kills) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_2_id AND mmtpp_with_date.date < m.date) AS team_2_player_2_avg_first_kills,
    (SELECT AVG(mmtpp_with_date.first_deaths) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_2_id AND mmtpp_with_date.date < m.date) AS team_2_player_2_avg_first_deaths,
    (SELECT AVG(mmtpp_with_date."2k" + mmtpp_with_date."3k" + mmtpp_with_date."4k" + mmtpp_with_date."5k") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_2_id AND mmtpp_with_date.date < m.date) AS team_2_player_2_avg_multikills,
    (SELECT AVG(mmtpp_with_date."1v1" + mmtpp_with_date."1v2" + mmtpp_with_date."1v3" + mmtpp_with_date."1v4" + mmtpp_with_date."1v5") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_2_id AND mmtpp_with_date.date < m.date) AS team_2_player_2_avg_clutches,
    -- NOTE: Statistic for team_2_player_3
    (SELECT AVG(mmtpp_with_date.rating) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_3_id AND mmtpp_with_date.date < m.date) AS team_2_player_3_avg_rating,
    (SELECT AVG(mmtpp_with_date.acs) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_3_id AND mmtpp_with_date.date < m.date) AS team_2_player_3_avg_acs,
    (SELECT AVG(mmtpp_with_date.kast) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_3_id AND mmtpp_with_date.date < m.date) AS team_2_player_3_avg_kast,
    (SELECT AVG(mmtpp_with_date.adr) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_3_id AND mmtpp_with_date.date < m.date) AS team_2_player_3_avg_adr,
    (SELECT AVG(mmtpp_with_date.first_kills) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_3_id AND mmtpp_with_date.date < m.date) AS team_2_player_3_avg_first_kills,
    (SELECT AVG(mmtpp_with_date.first_deaths) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_3_id AND mmtpp_with_date.date < m.date) AS team_2_player_3_avg_first_deaths,
    (SELECT AVG(mmtpp_with_date."2k" + mmtpp_with_date."3k" + mmtpp_with_date."4k" + mmtpp_with_date."5k") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_3_id AND mmtpp_with_date.date < m.date) AS team_2_player_3_avg_multikills,
    (SELECT AVG(mmtpp_with_date."1v1" + mmtpp_with_date."1v2" + mmtpp_with_date."1v3" + mmtpp_with_date."1v4" + mmtpp_with_date."1v5") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_3_id AND mmtpp_with_date.date < m.date) AS team_2_player_3_avg_clutches,
    -- NOTE: Statistic for team_2_player_4
    (SELECT AVG(mmtpp_with_date.rating) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_4_id AND mmtpp_with_date.date < m.date) AS team_2_player_4_avg_rating,
    (SELECT AVG(mmtpp_with_date.acs) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_4_id AND mmtpp_with_date.date < m.date) AS team_2_player_4_avg_acs,
    (SELECT AVG(mmtpp_with_date.kast) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_4_id AND mmtpp_with_date.date < m.date) AS team_2_player_4_avg_kast,
    (SELECT AVG(mmtpp_with_date.adr) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_4_id AND mmtpp_with_date.date < m.date) AS team_2_player_4_avg_adr,
    (SELECT AVG(mmtpp_with_date.first_kills) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_4_id AND mmtpp_with_date.date < m.date) AS team_2_player_4_avg_first_kills,
    (SELECT AVG(mmtpp_with_date.first_deaths) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_4_id AND mmtpp_with_date.date < m.date) AS team_2_player_4_avg_first_deaths,
    (SELECT AVG(mmtpp_with_date."2k" + mmtpp_with_date."3k" + mmtpp_with_date."4k" + mmtpp_with_date."5k") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_4_id AND mmtpp_with_date.date < m.date) AS team_2_player_4_avg_multikills,
    (SELECT AVG(mmtpp_with_date."1v1" + mmtpp_with_date."1v2" + mmtpp_with_date."1v3" + mmtpp_with_date."1v4" + mmtpp_with_date."1v5") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_4_id AND mmtpp_with_date.date < m.date) AS team_2_player_4_avg_clutches,
    -- NOTE: Statistic for team_2_player_5
    (SELECT AVG(mmtpp_with_date.rating) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_5_id AND mmtpp_with_date.date < m.date) AS team_2_player_5_avg_rating,
    (SELECT AVG(mmtpp_with_date.acs) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_5_id AND mmtpp_with_date.date < m.date) AS team_2_player_5_avg_acs,
    (SELECT AVG(mmtpp_with_date.kast) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_5_id AND mmtpp_with_date.date < m.date) AS team_2_player_5_avg_kast,
    (SELECT AVG(mmtpp_with_date.adr) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_5_id AND mmtpp_with_date.date < m.date) AS team_2_player_5_avg_adr,
    (SELECT AVG(mmtpp_with_date.first_kills) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_5_id AND mmtpp_with_date.date < m.date) AS team_2_player_5_avg_first_kills,
    (SELECT AVG(mmtpp_with_date.first_deaths) FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_5_id AND mmtpp_with_date.date < m.date) AS team_2_player_5_avg_first_deaths,
    (SELECT AVG(mmtpp_with_date."2k" + mmtpp_with_date."3k" + mmtpp_with_date."4k" + mmtpp_with_date."5k") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_5_id AND mmtpp_with_date.date < m.date) AS team_2_player_5_avg_multikills,
    (SELECT AVG(mmtpp_with_date."1v1" + mmtpp_with_date."1v2" + mmtpp_with_date."1v3" + mmtpp_with_date."1v4" + mmtpp_with_date."1v5") FROM mmtpp_with_date WHERE mmtpp_with_date.player_id = matches_players.team_2_player_5_id AND mmtpp_with_date.date < m.date) AS team_2_player_5_avg_clutches,
    -- NOTE: Team won (label)
    (CASE WHEN m.team_1_id = m.team_won THEN 1 ELSE 2 END) AS team_won_label
FROM matches m JOIN matches_players ON m.id = matches_players.match_id
ORDER BY m.date DESC;

