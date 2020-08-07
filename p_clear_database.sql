/*请在sys库上配置存储过程和定时器*/

/*存储过程定是清理测试库*/
DROP PROCEDURE  if exists p_clear_database;
delimiter $$
CREATE  PROCEDURE p_clear_database()
BEGIN
    DECLARE g_database VARCHAR(100);
    DECLARE done bit DEFAULT 0;
    DECLARE g_cursor CURSOR FOR select distinct(TABLE_SCHEMA) from `information_schema`.TABLES where TABLE_SCHEMA like 'test_%';
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = 1;
    SET @v_full_sql := '';
    OPEN g_cursor;
    REPEAT
        FETCH g_cursor into g_database;
            set @v_full_sql = CONCAT('drop database if exists ',g_database);
            PREPARE stmt from @v_full_sql;
            execute stmt;
            DEALLOCATE PREPARE stmt;
    UNTIL done END REPEAT;
    CLOSE g_cursor;
select 'OK';
END$$


/*定时任务*/
DROP EVENT IF EXISTS e_clear_database;
CREATE EVENT e_clear_database
ON SCHEDULE EVERY 1 DAY STARTS date_add(date( ADDDATE(curdate(),1)),interval 3 hour)
DO call p_clear_database();