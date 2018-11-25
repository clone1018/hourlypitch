CREATE TABLE `ideas` ( 
    `id` INTEGER PRIMARY KEY AUTOINCREMENT, 
    `pitch` TEXT, 
    `created` INTEGER, 
    `approved` INTEGER DEFAULT NULL,
    `shown` INTEGER DEFAULT NULL 
);