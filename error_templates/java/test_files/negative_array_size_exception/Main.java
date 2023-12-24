public class Main {
    public static void main(String[] args) {
        // Attempting to create an array with a negative size
        int[] array = new int[-5];
        System.out.println(array.length);
    }
}