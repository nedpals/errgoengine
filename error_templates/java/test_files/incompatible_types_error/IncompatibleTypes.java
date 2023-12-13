public class IncompatibleTypes {
    public static void main(String[] args) {
        int number = 10;
        String text = "Hello";

        // Incompatible types: Cannot convert int to String
        text = number;
    }
}
