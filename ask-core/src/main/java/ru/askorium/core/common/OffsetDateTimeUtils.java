package ru.askorium.core.common;

import lombok.AccessLevel;
import lombok.NoArgsConstructor;

import java.time.DayOfWeek;
import java.time.OffsetDateTime;

@NoArgsConstructor(access = AccessLevel.PRIVATE)
public class OffsetDateTimeUtils {

    public static OffsetDateTime getStartOfWeek(OffsetDateTime dateTime) {
        return getStartOfDay(dateTime)
                .with(DayOfWeek.MONDAY);
    }

    public static OffsetDateTime getStartOfDay(OffsetDateTime dateTime) {
        return dateTime
                .withHour(0)
                .withMinute(0)
                .withSecond(0)
                .withNano(0);
    }

}
