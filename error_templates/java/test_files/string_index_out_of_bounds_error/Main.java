public class Main {
    public static void main(String[] args) {
        String text = "Hello, World!";
        // Attempting to access an index beyond the string's length
        char character = text.charAt(20);
        System.out.println(character);
    }
}
