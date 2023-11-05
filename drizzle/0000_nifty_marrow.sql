CREATE TABLE `messages` (
	`messageId` integer PRIMARY KEY NOT NULL,
	`chatId` integer NOT NULL,
	`message` text NOT NULL,
	`messageDate` integer NOT NULL
);
