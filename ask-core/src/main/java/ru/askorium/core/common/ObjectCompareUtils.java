package ru.askorium.core.common;

import lombok.AccessLevel;
import lombok.NoArgsConstructor;

import java.util.Arrays;
import java.util.HashSet;
import java.util.Objects;

@NoArgsConstructor(access = AccessLevel.PRIVATE)
public class ObjectCompareUtils {

    public static boolean equalsObjectsExcludeFields(Object obj1, Object obj2, String... excludeFields) {
        if (obj1 == null || obj2 == null) {
            return false;
        }

        if (!obj1.getClass().equals(obj2.getClass())) {
            return false;
        }

        if (Objects.equals(obj1, obj2)) {
            return true;
        }

        var excludeSet = new HashSet<>(Arrays.asList(excludeFields));

        try {
            for (var field : obj1.getClass().getDeclaredFields()) {
                field.setAccessible(true);

                if (excludeSet.contains(field.getName())) {
                    continue;
                }

                var value1 = field.get(obj1);
                var value2 = field.get(obj2);

                if (!Objects.equals(value1, value2)) {
                    return false;
                }

                field.setAccessible(false);
            }

            return true;
        } catch (IllegalAccessException e) {
            throw new RuntimeException("Failed to compare objects", e);
        }
    }

}
