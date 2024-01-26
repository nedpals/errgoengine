import java.util.*;

public class Main {
    public static void main(String[] args) {
        List<String> myList = new ArrayList<>();
        Iterator<String> iterator = myList.iterator();
        // Attempting to get an element from an empty list
        String element = iterator.next();
        System.out.println(element);
    }
}
